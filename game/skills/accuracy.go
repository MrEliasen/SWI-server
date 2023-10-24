package skills

import "math/rand"

type Accuracy struct {
	Value float32
}

func (skill *Accuracy) SkillCheck() bool {
	v := float32(rand.Intn(10001)) / float32(100)
	success := v <= skill.Value
	skill.Train(success)

	return success
}

func (skill *Accuracy) Train(success bool) {
	var amount float32 = 0.0

	switch {
	case skill.Value < 20:
		amount = 1

	case skill.Value < 30:
		amount = 0.2

	case skill.Value < 40:
		amount = 0.08

	case skill.Value < 50:
		amount = 0.04

	case skill.Value < 60:
		amount = 0.008

	case skill.Value < 70:
		amount = 0.004

	case skill.Value < 80:
		amount = 0.0008

	case skill.Value < 90:
		amount = 0.0004

	case skill.Value < 100:
		amount = 0.00008

	default:
		amount = 0.0
	}

	if !success {
		if skill.Value > 35 {
			amount = 0.0
		} else {
			amount = amount / 15
		}
	}

	skill.Value += amount
}
