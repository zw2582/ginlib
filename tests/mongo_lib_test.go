package tests

import (
	"github.com/zw2582/ginlib"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
	"testing"
	"time"
	"context"
)

const COLLECTION_SHOP_ADDR = "location"

//ShopLocationIndex 创建地理位置索引
func TestCreateIndex(t *testing.T)  {
	ginlib.InitMongoDb()

	ginlib.Mongo(COLLECTION_SHOP_ADDR, func(collection *mongo.Collection) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		//创建唯一索引，位置索引
		name, err := collection.Indexes().CreateMany(ctx, []mongo.IndexModel{
			mongo.IndexModel{
				Keys:    bson.M{"shop_addr_id":1},
				Options:options.Index().SetUnique(true),
			},
			mongo.IndexModel{
				Keys:    bson.M{"loc":"2dsphere"},
			},
		})
		if err != nil {
			panic(err)
		}
		ginlib.Logger.Info("mongo 创建索引成功", zap.Any("index", name))
	})
}


//TestLocationAdd 新增位置
func TestLocationAdd(t *testing.T)  {
	ginlib.InitMongoDb()

	ginlib.Mongo(COLLECTION_SHOP_ADDR, func(collection *mongo.Collection) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		//插入经纬度
		shopId := 1
		long := 123.23112
		lati := 23.56
		collection.InsertOne(ctx, bson.M{
			"shop_id":shopId,
			"loc":bson.M{
				"type":"Point",
				"coordinates":[]float64{long, lati},
			},
		})
	})
}

//LocationDel 删除位置
func LocationDel(t *testing.T)  {
	ginlib.InitMongoDb()

	ginlib.Mongo(COLLECTION_SHOP_ADDR, func(collection *mongo.Collection) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		//删除经纬度
		shop_id := 1
		collection.DeleteOne(ctx, bson.M{
			"shop_id":shop_id,
		})
	})
}

