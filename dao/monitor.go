package dao

/**
 * 监控项&报警
 */
type Monitor struct {
	Id int
	itemName string
	Threshold string // 报警的阈值
	ThresholdNum string // 累计超过报警的阈值多少次后报警
	CurrentThresholdNum string // 当前累计超过报警的阈值次数（当监控项恢复正常后，这个值被摸为0）
	OldThresholdNum string // 历史累计超过报警的阈值次数（只会累加，不会被抹掉）
	Action string // 报警执行的动作
}

func NewMonitor(itemName string) *Monitor{
	return &Monitor{
		itemName: itemName,
	}
}


/**
 * 有则更新、无则插入监控项
 */
func(m *Monitor) SaveOrUpdate(){
	// 先查询，再修改
}