package filter

// strongRoots are blocked anywhere inside a word (unless the whole word is allow-listed).
var strongRoots = []string{
	"fuck", "fuk", "fuq", "phuck", "phuk", "shit", "cunt", "bitch", "nigg", "whore",
	"slut", "bastard", "porn", "pussy", "dildo", "cock", "penis", "vagina", "nazi",
	"hitler", "retard", "jizz", "twat", "wank", "asshole", "jackass", "dumbass",
	"damn", "piss", "faggot", "murder", "horny", "bollock",
}

// wordRoots are blocked only as whole words (plus common endings), because
// they hide inside lots of innocent words: class, title, skill, shell...
var wordRoots = []string{
	"ass", "arse", "dick", "tit", "anal", "sex", "boob", "butt", "crap", "hell",
	"rape", "rapist", "fag", "kill", "hate", "stupid", "idiot", "dumb", "loser",
	"moron", "sucks", "suck", "fart", "kkk", "poo", "pee",
}

// allowWords contain a strong root but are fine.
var allowWords = []string{
	"peacock", "cockatoo", "cockatoos", "cockatiel", "hancock", "hitchcock", "woodcock",
	"shuttlecock", "cocktail", "cockpit", "scunthorpe", "swank", "swanky", "shitake",
}
