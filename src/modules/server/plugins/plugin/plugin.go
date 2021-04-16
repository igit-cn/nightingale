package collector

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/didi/nightingale/v4/src/models"
	"github.com/didi/nightingale/v4/src/modules/server/collector"

	"github.com/influxdata/telegraf"
)

func init() {
	collector.CollectorRegister(&PluginCollector{})
}

type PluginCollector struct{}

func (p PluginCollector) Name() string                   { return "plugin" }
func (p PluginCollector) Category() collector.Category   { return collector.LocalCategory }
func (p PluginCollector) Template() (interface{}, error) { return nil, nil }
func (p PluginCollector) TelegrafInput(*models.CollectRule) (telegraf.Input, error) {
	return nil, errors.New("unsupported")
}

func (p PluginCollector) Get(id int64) (interface{}, error) {
	collect := new(models.PluginCollect)
	has, err := models.DB["mon"].Where("id = ?", id).Get(collect)
	if !has {
		return nil, err
	}
	return collect, err
}

func (p PluginCollector) Gets(nids []int64) (ret []interface{}, err error) {
	collects := []models.PluginCollect{}
	err = models.DB["mon"].In("nid", nids).Find(&collects)
	for _, c := range collects {
		ret = append(ret, c)
	}
	return ret, err
}

func (p PluginCollector) GetByNameAndNid(name string, nid int64) (interface{}, error) {
	collect := new(models.PluginCollect)
	has, err := models.DB["mon"].Where("name = ? and nid = ?", name, nid).Get(collect)
	if !has {
		return nil, err
	}
	return collect, err
}

func (p PluginCollector) Create(data []byte, username string) error {
	collect := new(models.PluginCollect)

	err := json.Unmarshal(data, collect)
	if err != nil {
		return fmt.Errorf("unmarshal body %s err:%v", string(data), err)
	}

	can, err := models.UsernameCandoNodeOp(username, "mon_collect_create", collect.Nid)
	if err != nil {
		return err
	}
	if !can {
		return fmt.Errorf("permission deny")
	}

	collect.Creator = username
	collect.LastUpdator = username

	nid := collect.Nid
	name := collect.Name

	old, err := p.GetByNameAndNid(name, nid)
	if err != nil {
		return err
	}
	if old != nil {
		return fmt.Errorf("同节点下策略名称 %s 已存在", name)
	}
	return models.CreateCollect(p.Name(), username, collect, false)
}

func (p PluginCollector) Update(data []byte, username string) error {
	collect := new(models.PluginCollect)

	err := json.Unmarshal(data, collect)
	if err != nil {
		return fmt.Errorf("unmarshal body %s err:%v", string(data), err)
	}

	can, err := models.UsernameCandoNodeOp(username, "mon_collect_modify", collect.Nid)
	if err != nil {
		return err
	}
	if !can {
		return fmt.Errorf("permission deny")
	}

	nid := collect.Nid
	name := collect.Name

	//校验采集是否存在
	obj, err := p.Get(collect.Id) //id找不到的情况
	if err != nil {
		return fmt.Errorf("采集不存在 type:%s id:%d", p.Name(), collect.Id)
	}

	tmpId := obj.(*models.PluginCollect).Id
	if tmpId == 0 {
		return fmt.Errorf("采集不存在 type:%s id:%d", p.Name(), collect.Id)
	}

	collect.Creator = username
	collect.LastUpdator = username

	old, err := p.GetByNameAndNid(name, nid)
	if err != nil {
		return err
	}
	if old != nil && tmpId != old.(*models.PluginCollect).Id {
		return fmt.Errorf("同节点下策略名称 %s 已存在", name)
	}

	return collect.Update()
}

func (p PluginCollector) Delete(id int64, username string) error {
	tmp, err := p.Get(id) //id找不到的情况
	if err != nil {
		return fmt.Errorf("采集不存在 type:%s id:%d", p.Name(), id)
	}
	nid := tmp.(*models.PluginCollect).Nid
	can, err := models.UsernameCandoNodeOp(username, "mon_collect_delete", int64(nid))
	if err != nil {
		return err
	}
	if !can {
		return fmt.Errorf("permission deny")
	}

	return models.DeleteCollectById(p.Name(), username, id)
}
