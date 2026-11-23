package redis

//redis key

// redis key注意尽量使用命名空间方式,区分不同得Key,方便查询和拆分

//定义为常量

const (
	KeyPrefix          = "bluebell:"
	KeyPostTimeZSet    = "post:time"   // zset:帖子及发帖时间
	KeyPostScoreZSet   = "post:score"  // zset:帖子及投票得分数
	KeyPostVotedZSetPF = "post:voted:" // zset;记录及投票用户得类型参数是post id
)

// 给redis Keyj加上前缀
func getRedisKey(key string) string {
	return KeyPrefix + key
}
