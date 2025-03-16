extends Control

# Called when the node enters the scene tree for the first time.
func _ready():
	pass

func _on_settings_button_pressed():
	get_tree().change_scene_to_file("res://settings.tscn")

func _on_exit_button_pressed():
	get_tree().quit()

func _on_live_pressed() -> void:
	get_tree().change_scene_to_file("res://live.tscn")
