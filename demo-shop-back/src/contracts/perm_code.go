package contracts

type PermCodesSource interface {
	GetPermCodesByUserId(userId int64) ([]string, error)
	GetPermCodesByApi(apiPath, requestMethod string) ([]string, error)
}
