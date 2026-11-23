package experiment

// Build a Robot (RW-2): snap on a head, arms, and movers, then test it on
// a challenge. Each challenge needs the right kind of part.

// Robot maps a slot to the part in it.
type Robot map[string]string

var RobotSlots = []string{"head", "arms", "movers"}

var RobotParts = map[string][]string{
	"head":   {"camera", "antenna", "sonar"},
	"arms":   {"claw", "magnet", "scoop"},
	"movers": {"wheels", "tracks", "legs"},
}

type RobotChallenge struct {
	ID     string
	Need   map[string]string // slot -> part that's required
	Expert bool
}

var robotChallenges = []RobotChallenge{
	{ID: "floor", Need: map[string]string{"movers": "wheels"}},
	{ID: "sand", Need: map[string]string{"movers": "tracks"}},
	{ID: "stairs", Need: map[string]string{"movers": "legs"}},
	{ID: "cave", Need: map[string]string{"head": "sonar", "movers": "legs"}, Expert: true},
	{ID: "junkyard", Need: map[string]string{"arms": "magnet", "movers": "tracks"}, Expert: true},
}

func RobotChallenges(expert bool) []RobotChallenge {
	var out []RobotChallenge
	for _, c := range robotChallenges {
		if expert || !c.Expert {
			out = append(out, c)
		}
	}
	return out
}

type RobotResult struct {
	OK        bool
	Discovery string
	Hint      string
	Boom      bool // the robot crashes in a funny way (E-4)
}

// Complete is true when every slot has a part.
func (r Robot) Complete() bool {
	for _, s := range RobotSlots {
		if r[s] == "" {
			return false
		}
	}
	return true
}

func TryRobot(r Robot, c RobotChallenge) RobotResult {
	if !r.Complete() {
		return RobotResult{Hint: "robot.hint.missing"}
	}
	for slot, part := range c.Need {
		if r[slot] != part {
			return RobotResult{Hint: "robot.hint." + c.ID, Boom: true}
		}
	}
	return RobotResult{OK: true, Discovery: "robot." + c.ID}
}
