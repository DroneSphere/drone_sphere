package service

type (
	WaylineSvc interface {
		//CreateWayline(points vo.)
	}

	WaylineRepo interface {
		SaveToS3(data []byte, key string) error
		FetchS3UrlByKey(key string) (string, error)
	}
)
