package file

import (
	"nurture/internal/config"
	"nurture/internal/file/handler"
	"nurture/internal/file/logic"

	"go.uber.org/zap"
)

type Deps struct {
	Config  config.Minio
	Storage logic.ObjectStorage
	Log     *zap.SugaredLogger
}

type Module struct {
	handler *handler.FileHandler
}

func NewModule(deps Deps) *Module {
	fileLogic := logic.NewFileLogic(deps.Storage, deps.Config, deps.Log)
	return &Module{
		handler: handler.NewFileHandler(fileLogic, deps.Log),
	}
}
