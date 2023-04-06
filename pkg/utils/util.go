package utils

import (
	"time"
)

var adminTeam = make([]string, 0)

var day2_unix_start int64 = 1645923600
var day2_unix_end int64 = 1645947000

func Init() {
	adminTeam = append(adminTeam, "team100")
	adminTeam = append(adminTeam, "team110")
	adminTeam = append(adminTeam, "team111")
}

func IsCompetitionTime(teamID string) bool {
	if IsAdminTeam(teamID) {
		return true
	}

	now := time.Now().Unix()

	if now <= day2_unix_start && now >= day2_unix_end {
		return false
	}

	return true
}

func IsPreparatoryPhase(teamID string) bool {
	if IsAdminTeam(teamID) {
		return false
	}
	now := time.Now().Unix()
	if now <= day2_unix_start || now >= day2_unix_end {
		return true
	}
	return false
}

func IsAdminTeam(teamId string) bool {
	for _, v := range adminTeam {
		if v == teamId {
			return true
		}
	}
	return false
}

func ToTimePtr(t time.Time) *time.Time {
	return &t
}
