package goass

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/Mericusta/go-stp"
)

type data_Logic struct {
	exp                       int64
	level                     int64
	dungeonProgress           int64
	attributesLevel           map[Attribution]int64
	skillsLevel               map[Skill]int64
	idleTS                    int64
	cumulativeRewards         int64
	saveDataPath              string
	generateProjectInputSlice []string
}

type json_Logic struct {
	Exp             int64                 `json:"exp"`
	Level           int64                 `json:"level"`
	DungeonProgress int64                 `json:"dungeon_progress"`
	AttributesLevel map[Attribution]int64 `json:"attributes_level"`
	SkillsLevel     map[Skill]int64       `json:"skill_level"`
	IdleTS          int64                 `json:"idle_ts"`
}

func newLogicData(saveDataPath string) *data_Logic {
	dataLogic := &data_Logic{
		exp:             0,
		level:           1,
		dungeonProgress: 1,
		attributesLevel: map[Attribution]int64{
			ATTRIBUTE_ATK: 0,
			ATTRIBUTE_DEF: 0,
			ATTRIBUTE_HP:  0,
			ATTRIBUTE_MP:  0,
		},
		skillsLevel: map[Skill]int64{
			SKILL_1: 0,
			SKILL_2: 0,
		},
		idleTS:       time.Now().Unix(),
		saveDataPath: saveDataPath,
	}
	dataLogic.LoadJSON()
	dataLogic.cumulativeRewards = time.Now().Unix() - dataLogic.idleTS
	return dataLogic
}

func (d *data_Logic) AddExp(delta int64) {
	d.exp += delta
	d.UpdateLevel()
}

func (d *data_Logic) UpdateLevel() {
	// 当前等级升级所需经验 = Σ(0, Level)
LEVEL_UP:
	levelUpNeedExp := int64(float64(0+d.level)*float64(d.level+1)/2) * 10
	if levelUpNeedExp > d.exp {
		return
	}
	d.level++
	d.exp -= levelUpNeedExp
	goto LEVEL_UP
}

func (d *data_Logic) IncreaseAttributeLevel(attributionKey Attribution) {
	if _, has := d.attributesLevel[attributionKey]; !has {
		panic(fmt.Sprintf("attribute key %v not exists", attributionKey))
	}
	release := d.level
	for _, level := range d.attributesLevel {
		release -= level
	}
	if release <= 0 {
		return
	}
	d.attributesLevel[attributionKey]++
}

func (d *data_Logic) IncreaseSkillLevel(skillKey Skill) {
	if _, has := d.skillsLevel[skillKey]; !has {
		panic(fmt.Sprintf("skill key %v not exists", skillKey))
	}
	release := d.level
	for _, level := range d.skillsLevel {
		release -= level
	}
	if release <= 0 {
		return
	}
	d.skillsLevel[skillKey]++
}

func (d *data_Logic) ToJSON() []byte {
	jsonBytes, err := json.MarshalIndent(&json_Logic{
		Exp:             d.exp,
		Level:           d.level,
		AttributesLevel: d.attributesLevel,
		SkillsLevel:     d.skillsLevel,
		IdleTS:          d.idleTS,
	}, "", "  ")
	if err != nil {
		panic(err)
	}

	return jsonBytes
}

func (d *data_Logic) LoadJSON() {
	if len(d.saveDataPath) == 0 {
		return
	}

	jsonLogic, err := stp.ReadFile(d.saveDataPath, func(b []byte) (*json_Logic, error) {
		if len(b) > 0 {
			jsonLogic := &json_Logic{}
			err := json.Unmarshal(b, jsonLogic)
			if err != nil {
				return nil, err
			}
			return jsonLogic, nil
		}
		return nil, nil
	})

	if err != nil {
		panic(err)
	}

	if jsonLogic == nil {
		return
	}

	if jsonLogic.Level == 0 || jsonLogic.IdleTS == 0 {
		panic("empty json data")
	}

	d.exp = jsonLogic.Exp
	d.level = jsonLogic.Level
	d.dungeonProgress = jsonLogic.DungeonProgress
	if d.dungeonProgress <= 0 {
		d.dungeonProgress = 1
	} else if d.dungeonProgress >= 10 {
		d.dungeonProgress = 10
	}
	for attribute, level := range jsonLogic.AttributesLevel {
		if _, has := d.attributesLevel[attribute]; has {
			d.attributesLevel[attribute] = level
		}
	}
	for skill, level := range jsonLogic.SkillsLevel {
		if _, has := d.skillsLevel[skill]; has {
			d.skillsLevel[skill] = level
		}
	}
	d.idleTS = jsonLogic.IdleTS
}
