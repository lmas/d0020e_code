extends Control

@onready var max_temperature = $SettingsContainer/maxTemp_text
@onready var min_temperature = $SettingsContainer/minTemp_text
@onready var max_price = $SettingsContainer/maxPrice_text
@onready var min_price = $SettingsContainer/minPrice_text
@onready var userTemp = $SettingsContainer/userTemp_text
@onready var Region = $SettingsContainer/Region_text
@onready var errorText = $ErrorText
@onready var tempError = 0
@onready var priceError = 0
@onready var regionError = 0

# TODO: Contact orchestrator to find comfortstat to get url instead.
var url = "http://130.240.110.106:8670/Comfortstat/Set_Values"

# Called when the node enters the scene tree for the first time.
func _ready():
	var newUrl = url + "/MaxTemperature"
	$"HTTP requests/get_maxTemp".request(newUrl)
	
	newUrl = url + "/MinTemperature"
	$"HTTP requests/get_minTemp".request(newUrl)
	
	newUrl = url + "/MaxPrice"
	$"HTTP requests/get_maxPrice".request(newUrl)
	
	newUrl = url + "/MinPrice"
	$"HTTP requests/get_minPrice".request(newUrl)
	
	newUrl = url + "/UserTemp"
	$"HTTP requests/get_userTemp".request(newUrl)
	
	newUrl = url + "/Region"
	$"HTTP requests/get_Region".request(newUrl)

func _on_url_text_text_changed(ip):
	self.url = "http://"+ip+":8670/Comfortstat/Set_Values"

func _on_get_max_temp_request_completed(_result, _response_code, _headers, body):
	var data = JSON.parse_string(body.get_string_from_utf8())
	max_temperature.text = str(data.value).pad_decimals(2)

func _on_get_min_temp_request_completed(_result, _response_code, _headers, body):
	var data = JSON.parse_string(body.get_string_from_utf8())
	min_temperature.text = str(data.value).pad_decimals(2)

func _on_get_max_price_request_completed(_result, _response_code, _headers, body):
	var data = JSON.parse_string(body.get_string_from_utf8())
	max_price.text = str(data.value).pad_decimals(2)

func _on_get_min_price_request_completed(_result, _response_code, _headers, body):
	var data = JSON.parse_string(body.get_string_from_utf8())
	min_price.text = str(data.value).pad_decimals(2)

func _on_get_user_temp_request_completed(_result, _response_code, _headers, body):
	var data = JSON.parse_string(body.get_string_from_utf8())
	userTemp.text = str(data.value).pad_decimals(1)

func _on_get_region_request_completed(_result, _response_code, _headers, body):
	var data = JSON.parse_string(body.get_string_from_utf8())
	Region.text = str(data.value).pad_decimals(0)

func _on_back_pressed():
	# Just goes back to main menu
	get_tree().change_scene_to_file("res://main_menu.tscn")

func _on_update_settings_pressed():
	if tempError or priceError or regionError:
		errorText.text = "Fix errors and try again."
		print("tempError:",tempError)
		print("priceError:", priceError)
		print("regionError:", regionError)
		return
	
	var http_request = HTTPRequest.new()
	add_child(http_request)
	http_request.request_completed.connect(self._http_request_completed)
	
	var payload = '{"value": %.1f, "version":"SignalA_v1.0"}' % max_temperature.text.to_float()
	var newUrl = url + "/MaxTemperature"
	var headers = ["Content-Type: application:json"]
	http_request.request(newUrl,headers,HTTPClient.METHOD_PUT,payload)
	await http_request.request_completed
	
	payload = '{"value": %.1f, "version":"SignalA_v1.0"}' % min_temperature.text.to_float()
	newUrl = url + "/MinTemperature"
	http_request.request(newUrl,headers,HTTPClient.METHOD_PUT,payload)
	await http_request.request_completed
	
	payload = '{"value": %.3f, "version":"SignalA_v1.0"}' % max_price.text.to_float()
	newUrl = url + "/MaxPrice"
	http_request.request(newUrl,headers,HTTPClient.METHOD_PUT,payload)
	await http_request.request_completed
	
	payload = '{"value": %.3f, "version":"SignalA_v1.0"}' % min_price.text.to_float()
	newUrl = url + "/MinPrice"
	http_request.request(newUrl,headers,HTTPClient.METHOD_PUT,payload)
	await http_request.request_completed
	
	payload = '{"value": %.3f, "version":"SignalA_v1.0"}' % userTemp.text.to_float()
	newUrl = url + "/UserTemp"
	http_request.request(newUrl,headers,HTTPClient.METHOD_PUT,payload)
	await http_request.request_completed
	
	payload = '{"value": %.0f, "version":"SignalA_v1.0"}' % Region.text.to_float()
	newUrl = url + "/Region"
	http_request.request(newUrl,headers,HTTPClient.METHOD_PUT,payload)
	await http_request.request_completed

 # Called when the HTTP request is completed.
func _http_request_completed(_result, _response_code, _headers, _body):
	pass

func _on_max_temp_text_changed(new_text):
	errorText.text = ""
	var red = Color(1.0,0.0,0.0,1.0)
	var white = Color(1.0,1.0,1.0,1.0)
	
	if float(new_text) > 30 or float(new_text) < 5:
		tempError = 1
		$SettingsContainer/maxTemp_text.add_theme_color_override("font_color",red)
		errorText.text = "Value out of range, range is 5 - 30 degrees"
	elif float(max_temperature.text) < float(min_temperature.text):
		tempError = 1
		$SettingsContainer/maxTemp_text.add_theme_color_override("font_color",red)
		errorText.text = "maxTemp must be higher than minTemp"
	elif new_text.contains(","):
		tempError = 1
		$SettingsContainer/maxTemp_text.add_theme_color_override("font_color",red)
		errorText.text = "Using wrong decimal symbol, please use '.'"
	else:
		tempError = 0
		$SettingsContainer/maxTemp_text.add_theme_color_override("font_color",white)
		$SettingsContainer/minTemp_text.add_theme_color_override("font_color",white)

func _on_min_temp_text_text_changed(new_text):
	errorText.text = ""
	var red = Color(1.0,0.0,0.0,1.0)
	var white = Color(1.0,1.0,1.0,1.0)
	
	if float(new_text) > 30 or float(new_text) < 5:
		tempError = 1
		$SettingsContainer/minTemp_text.add_theme_color_override("font_color",red)
		errorText.text = "Value out of range, range is 5 to maxTemp"
	elif float(min_temperature.text) > float(max_temperature.text):
		tempError = 1
		$SettingsContainer/minTemp_text.add_theme_color_override("font_color",red)
		errorText.text = "minTemp must be lower than maxTemp"
	elif new_text.contains(","):
		tempError = 1
		$SettingsContainer/minTemp_text.add_theme_color_override("font_color",red)
		errorText.text = "Using wrong decimal symbol, please use '.'"
	else:
		tempError = 0
		$SettingsContainer/minTemp_text.add_theme_color_override("font_color",white)
		$SettingsContainer/maxTemp_text.add_theme_color_override("font_color",white)

func _on_max_price_text_text_changed(new_text):
	errorText.text = ""
	var red = Color(1.0,0.0,0.0,1.0)
	var white = Color(1.0,1.0,1.0,1.0)
	
	if float(max_price.text) < float(min_price.text):
		priceError = 1
		$SettingsContainer/maxPrice_text.add_theme_color_override("font_color",red)
		errorText.text = "maxPrice has to be higher than minPrice"
	elif new_text.contains(","):
		priceError = 1
		$SettingsContainer/maxPrice_text.add_theme_color_override("font_color",red)
		errorText.text = "Using wrong decimal symbol, please use '.'"
	else:
		priceError = 0
		$SettingsContainer/maxPrice_text.add_theme_color_override("font_color",white)
		$SettingsContainer/minPrice_text.add_theme_color_override("font_color",white)

func _on_min_price_text_text_changed(new_text):
	errorText.text = ""
	var red = Color(1.0,0.0,0.0,1.0)
	var white = Color(1.0,1.0,1.0,1.0)
	
	if float(min_price.text) > float(max_price.text):
		priceError = 1
		$SettingsContainer/minPrice_text.add_theme_color_override("font_color",red)
		errorText.text = "minPrice must be lower than maxPrice"
	elif new_text.contains(","):
		priceError = 1
		$SettingsContainer/minPrice_text.add_theme_color_override("font_color",red)
		errorText.text = "Using wrong decimal symbol, please use '.'"
	else:
		priceError = 0
		$SettingsContainer/minPrice_text.add_theme_color_override("font_color",white)
		$SettingsContainer/maxPrice_text.add_theme_color_override("font_color",white)

func _on_user_temp_text_text_changed(new_text):
	errorText.text = ""
	var red = Color(1.0,0.0,0.0,1.0)
	var white = Color(1.0,1.0,1.0,1.0)
	
	if (float(userTemp.text) > 30 or float(userTemp.text) < 5) and float(userTemp.text) != 0.0:
		tempError = 1
		$SettingsContainer/userTemp_text.add_theme_color_override("font_color",red)
		errorText.text = "Value out of range, range is 5 to 30. Reset by setting to 0.0"
	elif new_text.contains(","):
		tempError = 1
		$SettingsContainer/userTemp_text.add_theme_color_override("font_color",red)
		errorText.text = "Using wrong decimal symbol, please use '.'"
	else:
		tempError = 0
		$SettingsContainer/userTemp_text.add_theme_color_override("font_color",white)
		$SettingsContainer/userTemp_text.add_theme_color_override("font_color",white)

func _on_region_text_text_changed(new_text: String) -> void:
	errorText.text = ""
	var red = Color(1.0,0.0,0.0,1.0)
	var white = Color(1.0,1.0,1.0,1.0)
	
	if float(Region.text) > 3 or float(Region.text) < 1:
		regionError = 1
		$SettingsContainer/Region_text.add_theme_color_override("font_color",red)
		errorText.text = "Region should be 1-3"
	elif new_text.contains(",") or new_text.contains("."):
		regionError = 1
		$SettingsContainer/Region_text.add_theme_color_override("font_color",red)
		errorText.text = "Please remove decimal points"
	else:
		regionError = 0
		$SettingsContainer/Region_text.add_theme_color_override("font_color",white)
		$SettingsContainer/Region_text.add_theme_color_override("font_color",white)
