package database

import (
	"blog-platform/internal/dto"
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	_ "github.com/joho/godotenv/autoload"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type BlogRepository interface {
	Health(ctx context.Context) error
	GetBlogs(ctx context.Context) ([]*Blog, error)
	GetBlog(ctx context.Context, id string) (*Blog, error)
	CreateBlog(ctx context.Context, create dto.BlogCreateDto) (*string, error)
	UpdateBlog(ctx context.Context, update dto.BlogUpdateDTO) (*Blog, error)
	DeleteBlog(ctx context.Context, id string) (*Blog, error)
	GetBlogsByTerm(ctx context.Context, term string) ([]*Blog, error)
}

type MongoBlogRepository struct {
	client     *mongo.Client
	collection *mongo.Collection
}

type Settings struct {
	HostName   string
	Username   string
	Password   string
	Port       string
	AuthSource string
	DbName     string
}

type Blog struct {
	ID        primitive.ObjectID `bson:"_id" json:"id"`
	CreatedAt time.Time          `bson:"created_at" json:"createdAt"`
	UpdatedAt time.Time          `bson:"updated_at" json:"updatedAt"`
	Title     string             `bson:"title" json:"title"`
	Category  string             `bson:"category" json:"category"`
	Content   string             `bson:"content" json:"content"`
	Tags      []string           `bson:"tags" json:"tags"`
}

func New(settings Settings) (*MongoBlogRepository, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	credential := options.Credential{
		AuthSource: settings.AuthSource,
		Username:   settings.Username,
		Password:   settings.Password,
	}

	uri := fmt.Sprintf("mongodb://%s:%s/", settings.HostName, settings.Port)

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri).SetAuth(credential))

	if err != nil {
		return nil, errors.New("database failed to connect")
	}

	return &MongoBlogRepository{
		client:     client,
		collection: client.Database(settings.DbName).Collection(settings.DbName),
	}, nil
}

func (s *MongoBlogRepository) Health(ctx context.Context) error {
	err := s.client.Ping(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to ping db - %v", err)
	}

	return nil
}

func (s *MongoBlogRepository) CreateBlog(ctx context.Context, create dto.BlogCreateDto) (*string, error) {
	blog := Blog{
		ID:        primitive.NewObjectID(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Title:     create.Title,
		Category:  create.Category,
		Content:   create.Content,
		Tags:      create.Tags,
	}

	result, err := s.collection.InsertOne(ctx, blog)
	if err != nil {
		return nil, fmt.Errorf("failed to insert blog entry - %v", err)
	}

	stringObjectID := result.InsertedID.(primitive.ObjectID).Hex()

	return &stringObjectID, nil
}

func (s *MongoBlogRepository) GetBlogs(ctx context.Context) ([]*Blog, error) {
	var blogs []*Blog
	filter := bson.D{{}}
	cur, err := s.collection.Find(ctx, filter)
	if err != nil {
		return nil, errors.New("no blogs found")
	}

	for cur.Next(ctx) {
		var b Blog
		err := cur.Decode(&b)
		if err != nil {
			return nil, errors.New("paging error")
		}
		blogs = append(blogs, &b)
	}

	if err := cur.Err(); err != nil {
		return nil, errors.New("paging error")
	}

	if err = cur.Close(ctx); err != nil {
		return nil, errors.New("paging error")
	}

	if len(blogs) == 0 {
		return nil, errors.New("no blogs found")
	}

	return blogs, nil
}

func (s *MongoBlogRepository) GetBlog(ctx context.Context, id string) (*Blog, error) {
	idFromHex, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, errors.New("failed to create new object id")
	}

	filter := bson.D{
		primitive.E{Key: "_id", Value: idFromHex}}

	var blog *Blog
	err = s.collection.FindOne(ctx, filter).Decode(&blog)
	if err != nil {
		fmt.Printf("cannot find that id %v", err)
	}

	return blog, nil
}

func (s *MongoBlogRepository) DeleteBlog(ctx context.Context, id string) (*Blog, error) {
	idFromHex, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		fmt.Println(err)
	}

	filter := bson.D{
		primitive.E{Key: "_id", Value: idFromHex}}

	var blog *Blog
	err = s.collection.FindOneAndDelete(ctx, filter).Decode(&blog)
	if err != nil {
		fmt.Printf("cannot find that id %v", err)
	}

	return blog, nil
}

func (s *MongoBlogRepository) UpdateBlog(ctx context.Context, update dto.BlogUpdateDTO) (*Blog, error) {
	objID, err := primitive.ObjectIDFromHex(update.Id)
	if err != nil {
		return nil, fmt.Errorf("invalid blog ID: %w", err)
	}

	updateFields := bson.M{}
	if update.Title != nil {
		updateFields["title"] = *update.Title
	}
	if update.Category != nil {
		updateFields["category"] = *update.Category
	}
	if update.Content != nil {
		updateFields["content"] = *update.Content
	}
	if update.Tags != nil {
		updateFields["tags"] = *update.Tags
	}
	updateFields["updated_at"] = time.Now()

	if len(updateFields) == 1 {
		return nil, fmt.Errorf("no valid fields to update")
	}

	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)

	var updated Blog
	err = s.collection.FindOneAndUpdate(
		ctx,
		bson.M{"_id": objID},
		bson.M{"$set": updateFields},
		opts,
	).Decode(&updated)

	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, fmt.Errorf("blog not found")
		}
		return nil, fmt.Errorf("failed to update blog: %w", err)
	}

	return &updated, nil
}

func (s *MongoBlogRepository) GetBlogsByTerm(ctx context.Context, term string) ([]*Blog, error) {
	var blogs []*Blog
	filter := bson.M{
		"$or": []bson.M{
			{"title": bson.M{"$regex": term, "$options": "i"}},
			{"content": bson.M{"$regex": term, "$options": "i"}},
			{"category": bson.M{"$regex": term, "$options": "i"}},
		},
	}
	cur, err := s.collection.Find(ctx, filter)
	if err != nil {
		log.Fatalf("db down: %v", err)
	}

	for cur.Next(ctx) {
		var b Blog
		err := cur.Decode(&b)
		if err != nil {
			return nil, errors.New("paging error")
		}
		blogs = append(blogs, &b)
	}

	if err := cur.Err(); err != nil {
		return nil, errors.New("paging error")
	}

	if err = cur.Close(ctx); err != nil {
		return nil, errors.New("paging error")
	}

	if len(blogs) == 0 {
		return nil, errors.New("no blogs")
	}

	fmt.Println(blogs)

	return blogs, nil
}
