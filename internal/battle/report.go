package battle

import (
	"encoding/json"
	"fmt"
)

// BattleReport 战报数据 (用于序列化)
type BattleReport struct {
	BattleID    int64            `json:"battle_id"`
	BattleType  string           `json:"battle_type"`
	Winner      string           `json:"winner"`
	Duration    int              `json:"duration"`
	StartTime   int64            `json:"start_time"`
	EndTime     int64            `json:"end_time"`
	Attacker    *SideReport      `json:"attacker"`
	Defender    *SideReport      `json:"defender"`
	Rewards     *BattleRewards   `json:"rewards,omitempty"`
}

// SideReport 战斗方战报
type SideReport struct {
	ID           int64            `json:"id"`
	Type         SideType         `json:"type"`
	HeroID       int64            `json:"hero_id,omitempty"`
	Power        int64            `json:"power"`
	Soldiers     map[string]int   `json:"soldiers"`
	Casualties   *CasualtyReport  `json:"casualties"`
}

// CasualtyReport 伤亡战报
type CasualtyReport struct {
	Death        map[string]int   `json:"death"`
	SeriousWound map[string]int   `json:"serious_wound"`
	MinorWound   map[string]int   `json:"minor_wound"`
	TotalDeath   int              `json:"total_death"`
	TotalSerious int              `json:"total_serious"`
	TotalMinor   int              `json:"total_minor"`
}

// GenerateReport 生成战报
func GenerateReport(result *BattleResult) *BattleReport {
	report := &BattleReport{
		BattleID:   result.ID,
		BattleType: result.Type.String(),
		Winner:     result.Winner,
		Duration:   result.Duration,
		StartTime:  result.StartTime,
		EndTime:    result.EndTime,
		Attacker:   generateSideReport(result.Attacker),
		Defender:   generateSideReport(result.Defender),
		Rewards:    result.Rewards,
	}

	return report
}

// generateSideReport 生成战斗方战报
func generateSideReport(side *BattleSide) *SideReport {
	return &SideReport{
		ID:         side.ID,
		Type:       side.Type,
		HeroID:     side.HeroID,
		Power:      side.Power,
		Soldiers:   convertSoldierMap(side.Soldiers),
		Casualties: generateCasualtyReport(side),
	}
}

// generateCasualtyReport 生成伤亡战报
func generateCasualtyReport(side *BattleSide) *CasualtyReport {
	return &CasualtyReport{
		Death:        convertSoldierMap(side.Death),
		SeriousWound: convertSoldierMap(side.SeriousWound),
		MinorWound:   convertSoldierMap(side.MinorWound),
		TotalDeath:   side.GetTotalDeaths(),
		TotalSerious: side.GetTotalSeriousWound(),
		TotalMinor:   side.GetTotalMinorWound(),
	}
}

// convertSoldierMap 转换士兵map (int key -> string key for JSON)
func convertSoldierMap(m map[int]int) map[string]int {
	result := make(map[string]int)
	for k, v := range m {
		result[fmt.Sprintf("%d", k)] = v
	}
	return result
}

// ToJSON 转换为JSON
func (r *BattleReport) ToJSON() ([]byte, error) {
	return json.Marshal(r)
}

// ToJSONString 转换为JSON字符串
func (r *BattleReport) ToJSONString() (string, error) {
	bytes, err := r.ToJSON()
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// Summary 生成战报摘要
func (r *BattleReport) Summary() string {
	winner := "防守方"
	if r.Winner == "attacker" {
		winner = "攻击方"
	}

	attackerLoss := r.Attacker.Casualties.TotalDeath + r.Attacker.Casualties.TotalSerious
	defenderLoss := r.Defender.Casualties.TotalDeath + r.Defender.Casualties.TotalSerious

	return fmt.Sprintf("战斗类型: %s, 胜者: %s, 攻击方伤亡: %d, 防守方伤亡: %d",
		r.BattleType, winner, attackerLoss, defenderLoss)
}
