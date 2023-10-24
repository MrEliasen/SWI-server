package skills

import "math/rand"

type Track struct {
	Value float32
}

func (skill *Track) SkillCheck() bool {
	v := float32(rand.Intn(10001)) / float32(100)
	success := v <= skill.Value

	if success {
		skill.Train()
	}

	return success
}

func (skill *Track) Train() {
	switch {
	case skill.Value < 10:
		skill.Value += 1

	case skill.Value < 20:
		skill.Value += 0.8

	case skill.Value < 30:
		skill.Value += 0.4

	case skill.Value < 40:
		skill.Value += 0.05

	case skill.Value < 50:
		skill.Value += 0.02

	case skill.Value < 60:
		skill.Value += 0.008

	case skill.Value < 70:
		skill.Value += 0.005

	case skill.Value < 80:
		skill.Value += 0.002

	case skill.Value < 90:
		skill.Value += 0.0008

	case skill.Value < 100:
		skill.Value += 0.0005

	default:
		skill.Value += 0.0
	}
}
