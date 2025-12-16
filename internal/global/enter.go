package global

import (
	"nurture/internal/pkg/zapx"

	"go.uber.org/zap"
)

var (
	Log *zap.SugaredLogger
)

func Init() {
	Log = zapx.InitZap()
}
