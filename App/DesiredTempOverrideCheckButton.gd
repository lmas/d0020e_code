extends CheckButton


# Called when the node enters the scene tree for the first time.
func _ready():
	pass # Replace with function body.


# Called every frame. 'delta' is the elapsed time since the previous frame.
func _process(delta):
	pass

func _on_toggled(button_pressed):
	if button_pressed:
		!get_tree().current_sccene.override
		print("Override:", get_tree().current_scene.override)
