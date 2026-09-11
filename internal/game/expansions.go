package game

import "fmt"

// ExpansionConfig is public, persisted, and locked once setup begins.
// Empty fields preserve the rules of pre-expansion EE saves.
type ExpansionConfig struct {
	Board          string `json:"board"`
	Structures     bool   `json:"structures"`
	SpecialWorkers bool   `json:"specialWorkers"`
	Visitors       string `json:"visitors"`
}

func (c ExpansionConfig) Normalized() ExpansionConfig {
	if c.Board == "" {
		c.Board = "ee"
	}
	if c.Visitors == "" {
		c.Visitors = "ee"
	}
	return c
}

func (c ExpansionConfig) Validate() error {
	c = c.Normalized()
	if c.Board != "ee" && c.Board != "tuscany" {
		return fmt.Errorf("无效主板：请选择 EE 或 Tuscany")
	}
	if c.Visitors != "ee" && c.Visitors != "ee_moor" && c.Visitors != "rhine" {
		return fmt.Errorf("无效访客牌组：EE、EE+Moor、Rhine 三选一；Rhine 不可混洗")
	}
	return nil
}

// A released configuration must contain only implemented visitor effects.
// Missing effects fail closed for both lobby changes and restored games.
func (c ExpansionConfig) Playable() error {
	if err := c.Validate(); err != nil {
		return err
	}
	c = c.Normalized()
	for _, card := range expansionCatalog(c) {
		if (card.Type == "summer" || card.Type == "winter") && !card.Implemented {
			return fmt.Errorf("访客 %s 尚未完成实现，暂不可开局", card.ID)
		}
	}
	return nil
}

func (r *Room) configure(id string, a Action) error {
	if r.Phase != "lobby" || r.HostID != id {
		return fmt.Errorf("只有房主可在大厅配置扩展")
	}
	if a.Revision != r.Revision {
		return fmt.Errorf("局面已经变化，请刷新操作")
	}
	if a.Config == nil {
		return fmt.Errorf("缺少完整开局配置")
	}
	c := a.Config.Normalized()
	if err := c.Playable(); err != nil {
		return err
	}
	r.Config = c
	r.AddLog("房主更新开局配置")
	return nil
}
