package builder

import "github.com/go-redis/redis"

func (b *Builder) SetBuildStatus(imageName string, data BuildStatusData) error {
	return b.redisClient.Set(imageName, data, defaultRedisMsgTTL).Err()
}

func (b *Builder) UpdateBuildStatus(imageName string, data BuildStatusData) error {
	bd, err := b.GetBuildStatus(imageName)
	if err != nil {
		return err
	}

	if !data.EndTime.IsZero() {
		bd.EndTime = data.EndTime
	}

	if data.Status != 0 {
		bd.Status = data.Status
	}

	bd.ErrorMsg = data.ErrorMsg
	bd.Logs += data.Logs
	return b.SetBuildStatus(imageName, bd)
}

func (b *Builder) GetBuildStatus(imageName string) (BuildStatusData, error) {
	var data BuildStatusData
	err := b.redisClient.Get(imageName).Scan(&data)
	if err != nil {
		if err == redis.Nil {
			return BuildStatusData{}, ErrBuildNotFound
		}
		return BuildStatusData{}, err
	}
	return data, nil
}
