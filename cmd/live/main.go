package main

import (
	"log"

	"shortvideo/internal/live/dao"
	"shortvideo/internal/live/handler"
	"shortvideo/internal/live/service"
	live "shortvideo/kitex_gen/live/liveservice"
	"shortvideo/pkg/database"
	"shortvideo/pkg/logger"
)

func main() {
	//初始化数据库连接
	db, err := database.Init()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	//初始化直播相关dao
	roomRepo := dao.NewLiveRoomRepository(db)
	giftRepo := dao.NewGiftRepository(db)
	giftRecordRepo := dao.NewGiftRecordRepository(db)
	roomAdminRepo := dao.NewRoomAdminRepository(db)
	liveRecordRepo := dao.NewLiveRecordRepository(db)
	roomViewerRepo := dao.NewRoomViewerRepository(db)

	//初始化直播服务
	liveService := service.NewLiveService(
		roomRepo,
		giftRepo,
		giftRecordRepo,
		roomAdminRepo,
		liveRecordRepo,
		roomViewerRepo,
	)

	//初始化处理器
	liveHandler := handler.NewLiveService(liveService)

	// 创建服务器
	svr := live.NewServer(liveHandler)

	// 启动服务器
	err = svr.Run()
	if err != nil {
		logger.Error("Failed to start live server", logger.ErrorField(err))
		log.Println(err.Error())
	}

	logger.Info("Live server started successfully")
}
