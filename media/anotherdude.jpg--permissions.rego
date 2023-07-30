package gosqlite

default Label = "TOP SECRET//MANHATTAN"
default LabelBg = "yellow"
default LabelFg = "black"
default Read = false
default Write = false
Read {
    input["age"][_] == "adult"
}
Write {
    input["email"][_] == "rob.fielding@gmail.com"
}
