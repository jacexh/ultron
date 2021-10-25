package ultron

type (
	SlaveSupervisor interface {
		Kill(string)
		StartNewPlan()
		FinishCurrentPlan()
		Send(AttackStrategy, Timer)
		Imprison(Slave)
	}

	Slave interface {
		ID() string
	}
)
