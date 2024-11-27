package storage

// 对应数据库中的行
type DataCell struct {
	Data map[string]interface{}
}

func (d *DataCell) GetTableName() string {
	return d.Data["Task"].(string)
}
func (d *DataCell) GetTaskName() string {
	return d.Data["Task"].(string)
}

// 存储接口
type Storage interface {
	Save(datas ...*DataCell) error
}
