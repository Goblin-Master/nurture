package repo

func (ur *UserRepo) logCacheHit(args ...interface{}) {
	ur.log.Info(args...)
}
