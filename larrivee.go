package main

var Larrivee = Levels{
	Gradient: []float64{10, 20, 25, 35, 42, 55, 70, 85, 90, 100},
	Risk:     []Danger{Extreme, Severe, High, Elevated, Moderate, Low, Moderate, Elevated, High, Severe},
	Details: &Notification{
		SubjectTmpl: `Your guitar is in {{.Risk}} danger!`,
		BodyTmpl: `Your guitar is not in the correct humidity range (42–55%).
You should make changes to the environment of the guitar to maintain this range.

Currently, the environment is in the range of {{.Low}}–{{.High}} relative humidity.
This humidity range is categorized as:

                                    {{.Risk}}

If the guitar remains in this humidity range, you can expect the following effects:

## 1–3 Days

{{.ShorttermEffects}}

## 3+ Days

{{.LongtermEffects}}

This information is made available by Larrivee.
See http://www.larivee.com/pdfs/Larrivee%20Care%20%20Maintenance.pdf for more information.


## Humidity Trend

You can use the following humidity trend chart to diagnose the problem.

{{.Trend}}

Greetings,

Ben Morgan
`,
		ShortEffects: []string{
			`Frets will feel sharp, top and back will become collapsed (concave), Action
will lower very quickly, and guitar will develop a buzz. Soundboard may develop
cracks especially running from the bridge to the butt, bridge may shear off.`,
			`Frets will likely feel sharp, top and back will likely become collapsed
(concave), Action will lower very quickly, guitar will likely develop a buzz.
Soundboard may develop a crack running from the bridge to the butt, bridge may
come unglued.`,
			`Frets will feel sharp, top may begin to  collapse (concave), Action will lower
very quickly, guitar may develop a buzz.`,
			`Fret ends may start to feel sharp, top may become slightly collapsed (concave).`,
			`No major problems should occur with limited exposure.`,
			`No problem will occur in this range.`,
			`No major problems should occur with limited exposure.`,
			`Sound quality may be diminished. Soundboard may appear swollen. Action may be
slightly high.`,
			`Guitar body may appear swollen, sound quality will slightly diminish,
playability may decrease. Action will become higher.`,
			`Guitar body will appear swollen and sound quality will be diminished,
playability will decrease. Back braces may come unglued as the wood expands.
Action will get very high quickly.`,
		},
		LongEffects: []string{
			`Fret Ends will feel very sharp, Soundboard and back will become collapsed
(concave), last six frets of the fingerboard will sink into sound hole, the
action will be extremely low with buzzes up and down the fingerboard, the
Bridge wings will appear concave, cracks will develop in the soundboard
especially from the bridge to the butt of the instrument, and the bridge will
shear off. Rosette rings and tail wedge will appear raised. Braces which do not
shear off may push out the binding of the instrument Braces will be visible as
high spots on the top and back.`,
			`Fret Ends will feel very sharp, Soundboard and back will be collapsed
(concave), last six frets of the fingerboard will sink into sound hole, action
will be lower, and guitar will buzz, Bridge wings will appear concave, a large
crack in the soundboard will likely develop from the bridge to the butt of the
instrument, and the bridge may come unglued.  Rosette rings and tail wedge may
be visibly raised.`,
			`Fret Ends will likely feel very sharp, Soundboard and back will become flat or
collapsed (concave), last six frets of the fingerboard will likely sink into
sound hole, the action will lower, and guitar will buzz, the Bridge wings will
appear concave, cracks in the soundboard may develop especially from the bridge
to the butt of the instrument, the bridge may shear off (come unglued).`,
			`Fret Ends will feel sharp, Soundboard and back will become flat or collapsed
(concave), action will feel lower, guitar will likely buzz, Bridge wings will
appear concave, after several months the bridge may “lift” or shear off.`,
			`Fret ends may feel sharp, soundboard may appear slightly collapsed (concave),
action may lower slightly, bridge wings will appear concave, the guitar may
develop a buzz.`,
			`No problem will occur in this range.`,
			`Top and back will appear bellied (convex), playability will be affected.
Fretboard from the 14th on may appear raised. Guitar will start to have a musty
smell after a couple of months.`,
			`Top and back will appear bellied (convex), playability will be affected.
Fretboard from the 14th on may appear raised. Guitar will start to have a musty
smell after a couple of months.`,
			`Braces will come loose after a few weeks, top and back will appear very bellied
(convex), bridge may loosen or come off, playability will be affected.
Fretboard from the 14th on will appear raised.  Mildew may form inside the
guitar.`,
			`All Glue Joints will loosen, Top and Back braces will loosen. Bridge may shear
come off the top, the guitar top will expand and belly (become convex) both in
front of and behind bridge, if the glue joints do not delaminate then the
guitar will be unplayable.  Fretboard from the 14th on will appear raised.
Mildew may form inside the guitar. Guitar will very likely de-construct itself.`,
		},
	},
}
