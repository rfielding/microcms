package gosqlite

default Label = "SECRET//SQUIRREL"
default LabelBg = "red"
default LabelFg = "white"
default Read = false
default Write = false
Read {
    input["age"][_] == "adult"
}
Write {
    input["email"][_] == "rob.fielding@gmail.com"
}