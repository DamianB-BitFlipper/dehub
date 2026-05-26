package actionview

import (
	"time"

	"charm.land/bubbles/v2/spinner"
)

const (
	largePaneWidth = 24
)

const (
	AsciiSkippedIcon = `
    ,---_   
   /   ╱ \  
  (   ╱   ) 
   \ ╱   /  
		:---:   
		`

	AsciiStoppedIcon = `
    .---.   
   /  │  \  
  |   │   | 
   \  .  /  
		:---:
		`

	Separator    = "│"
	ExpandSymbol = "▶"
	ListSymbol   = "≡"
	Ellipsis     = "…"
)

const emptySetIllustration = "            ╱   \n" +
	"    ..-··-.╱    \n" +
	"  .´      ╱ `.  \n" +
	" /       ╱    \\ \n" +
	".       ╱      .\n" +
	":      ╱       :\n" +
	"'     ╱        '\n" +
	" \\   ╱        / \n" +
	"  `.╱       .´  \n" +
	"   ╱`·-..-·´    \n" +
	"  ╱             \n"

const stopSignArt = "    _________    \n" +
	"   /         \\   \n" +
	"  /    .-.    \\  \n" +
	" /     | |     \\ \n" +
	"|      | |      |\n" +
	"|      | |      |\n" +
	"|      !_!      |\n" +
	" \\             / \n" +
	"  \\     O     /  \n" +
	"   \\         /   \n" +
	"    ‾‾‾‾‾‾‾‾‾    \n"

const checkmarkSignArt = "    ..-··-..    \n" +
	"  .´        `.  \n" +
	" /        /   \\ \n" +
	".        /     .\n" +
	":    \\  /      :\n" +
	"'     \\/       '\n" +
	" \\            / \n" +
	"  `.        .´  \n" +
	"    `·-..-·´    \n" +
	"                \n"

var LogsFrames = spinner.Spinner{
	Frames: []string{
		`
 ╭─────────╮   
 │FEDGHHFUR│   
 │ORUDKFMVR│   
 │NFEYNFDSN│   
 │NFEYFNWYF│   
 │MADODJWJF│   
 │FHUEHFISI│──╮
 │YFURKSIFJ│╭╮│
 │UDYGJDIUW│─╯│
 ╰─────────╰──╯
               
  Loading.     
`,
		`
 ╭─────────╮   
 │-1101100-│   
 │-1101101-│   
 │-1100001-│   
 │-1101111-│   
 │-1101100-│   
 │-1101111-│──╮
 │-1101100-│╭╮│
 │-1100101-│─╯│
 ╰─────────╰──╯
               
  Loading.     
`,
		`
 ╭─────────╮   
 │^№№*%^)?:│   
 │)(&№:?@!~│   
 │/\'"[]{&$│   
 │$^()%&^$#│   
 │#$%"%;&^&│   
 │^&^??\/\"│──╮
 │^%%*&#()$│╭╮│
 │(?*;?%%^&│─╯│
 ╰─────────╰──╯
               
  Loading.     
`,
		`
 ╭─────────╮   
 │654130037│   
 │103985647│   
 │376247259│   
 │184537563│   
 │184764464│   
 │104749275│──╮
 │367858324│╭╮│
 │438756456│─╯│
 ╰─────────╰──╯
               
  Loading.     
`,
		`
 ╭─────────╮   
 │-1101100-│   
 │-1101101-│   
 │-1100001-│   
 │-1101111-│   
 │-1101100-│   
 │-1101111-│──╮
 │-1101100-│╭╮│
 │-1100101-│─╯│
 ╰─────────╰──╯
               
  Loading..    
`,
		`
 ╭─────────╮   
 │FEDGHHFUR│   
 │ORUDKFMVR│   
 │NFEYNFDSN│   
 │NFEYFNWYF│   
 │MADODJWJF│   
 │FHUEHFISI│──╮
 │YFURKSIFJ│╭╮│
 │UDYGJDIUW│─╯│
 ╰─────────╰──╯
               
  Loading..    
`,
		`
 ╭─────────╮   
 │-+-+-+-+-│   
 │-+-+-+-+-│   
 │-+-+-+-+-│   
 │-+-+-+-+-│   
 │-+-+-+-+-│   
 │-+-+-+-+-│──╮
 │-+-+-+-+-│╭╮│
 │-+-+-+-+-│─╯│
 ╰─────────╰──╯
               
  Loading...   
`,
		`
 ╭─────────╮   
 │^№№*%^)?:│   
 │)(&№:?@!~│   
 │/\'"[]{&$│   
 │$^()%&^$#│   
 │#$%"%;&^&│   
 │^&^??\\/"│──╮
 │^%%*&#()$│╭╮│
 │(?*;?%%^&│─╯│
 ╰─────────╰──╯
               
  Loading...   
`,
		`
 ╭─────────╮   
 │654130037│   
 │103985647│   
 │376247259│   
 │184537563│   
 │184764464│   
 │104749275│──╮
 │367858324│╭╮│
 │438756456│─╯│
 ╰─────────╰──╯
               
  Loading...   
`,
	},
	FPS: time.Second / 10,
}

var InProgressFrames = spinner.Spinner{
	Frames: []string{
		`
  ▀▀ 
     
     
`,

		`
   ▀▜
     
     
`,
		`
    ▜
    ▐
     
`,
		`
     
    ▐
    ▟
`,
		`
     
     
   ▄▟
`,
		`
     
     
  ▄▄ 
`,
		`
     
     
 ▄▄  
`,
		`
     
     
▙▄   
`,
		`
     
▌    
▙    
`,
		`
▛    
▌    
     
`,
		`
▛▀   
     
     
`,
		`
 ▀▀  
     
     
`,
	},
	FPS: time.Second / 12,
}

var ClockFrames = spinner.Spinner{
	Frames: []string{"󱑌", "󱑍", "󱑎", "󱑏", "󱑐", "󱑑", "󱑒", "󱑓", "󱑔", "󱑕", "󱑖", "󱑋"},
	FPS:    time.Second / 6,
}

var SpinnerFrames = spinner.Spinner{
	Frames: []string{"󰪞", "󰪟", "󰪠", "󰪡", "󰪢", "󰪣", "󰪤", "󰪥"},
	FPS:    time.Second / 6,
}

var MoonSpinnerFrames = spinner.Spinner{
	Frames: []string{
		"", "", "", "", "", "", "", "", "", "", "",
		"", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "",
	},
	FPS: time.Second / 12,
}
