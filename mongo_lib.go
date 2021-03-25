package ginlib

import (
	"context"
	"fmt"
	pool "github.com/jolestar/go-commons-pool"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"go.uber.org/zap"
	"time"
)

const (
	COLLECTION_SHOP_ADDR = "shop_addr" //店铺位置
)

var (
	mongoPool *pool.ObjectPool
)

type MongoDbFunc func(collection *mongo.Collection)

//初始化mongodb
func InitMongoDb()  {
	mongoHost := Ini_Str("mongo.host", "localhost")
	mongoPort := Ini_Str("mongo.port", "27017")
	factory := pool.NewPooledObjectFactory(func(ctx context.Context) (i interface{}, err error) {
		//create
		Logger.Info("创建mongodb连接实例")

		ctx, cannel := context.WithTimeout(context.Background(), time.Second*10)
		defer cannel()
		//连接mongodb
		client, err := mongo.Connect(ctx, options.Client().ApplyURI(fmt.Sprintf("mongodb://%s:%s", mongoHost, mongoPort)))
		return client, err
	}, func(ctx context.Context, object *pool.PooledObject) error {
		//destroy
		Logger.Info("销毁mongodb连接实例")

		ctx, cannel := context.WithTimeout(context.Background(), time.Second*2)
		defer cannel()

		client := object.Object.(*mongo.Client)
		return client.Disconnect(ctx)
	}, func(ctx context.Context, object *pool.PooledObject) bool {
		//validate
		Logger.Info("校验mongodb连接实例是否有效")

		ctx, cannel := context.WithTimeout(context.Background(), time.Second*2)
		defer cannel()

		client := object.Object.(*mongo.Client)
		if err := client.Ping(ctx, readpref.Primary()); err != nil {
			Logger.Error("mongodb ping fail:"+err.Error(), zap.Error(err))
			return false
		}
		return true
	}, func(ctx context.Context, object *pool.PooledObject) error {
		//activate
		return nil
	}, func(ctx context.Context, object *pool.PooledObject) error {
		//passivate
		return nil
	})

	mongoPool = pool.NewObjectPoolWithDefaultConfig(context.Background(), factory)
}

func Mongo(collectionName string, fun MongoDbFunc)  {
	//获取mongodb连接
	obj, err := mongoPool.BorrowObject(context.Background())
	if err != nil {
		panic(err)
	}
	//默认归还
	defer func() {
		if err := mongoPool.ReturnObject(context.Background(), obj); err != nil {
			panic(err)
		}
	}()

	//转换连接
	client := obj.(*mongo.Client)

	//选择数据库
	database := APP_NAME
	collection := client.Database(database).Collection(collectionName)

	//执行操作
	fun(collection)

}

