package microcms
default Label = "ADMIN"
default LabelBg = "blue"
default LabelFg = "white"	
default Read = true
default Write = false
Write {
	input["role"][_] == "admin"
}
