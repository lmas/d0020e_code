[![Tests and Linters](https://github.com/lmas/d0020e_code/actions/workflows/main.yml/badge.svg?branch=master)](https://github.com/lmas/d0020e_code/actions/workflows/main.yml)

# D0020E, Projekt i datateknik, H24

"Demand Response with ZigBee and mbaigo."

**The public code repo.**

## Description

This project aims to reduce society's energy consumption by controlling smart power
devices, making adjustments according to the daily power market and the local weather.
Powered by a Internet-of-Things cloud that you can run locally and privately,
using an alternative [Arrowhead] implementation in [Go].

[Arrowhead]: https://arrowhead.eu/eclipse-arrowhead-2/
[Go]: https://github.com/sdoque/mbaigo



# Hardware setup

An initial hardware setup is required before being able to run the systems and devices.
Following documents how to set up a Raspberry Pi as a host for the cloud.

## Step 1: Flashing a Raspberry Pi OS to the Raspberry Pi memory card

- Connect the Micro SD card into your computer.

- Download Raspberry Pi OS using the Raspberry Pi Imager downloaded at, for Raspberry Pi Connect to work the Raspberry Pi has to run Raspberry Pi OS bookworm or later: https://www.raspberrypi.com/software/

- Follow the instruction and wait until the OS is loaded onto the SD card.

## Step 2: Booting up the Raspberry Pi

- Insert the SD card into the Raspberry Pi, connect it to power and connect it to a screen with a micro-HDMI to HDMI cable.

- When the Raspberry Pi has started, connect a keyboard and mouse to it.

- Either connect the Raspberry Pi to internet with an Ethernet cable or in the top right corner connect the Raspberry Pi to a wireless internet.

## Step 3: Install Raspberry Pi Connect on the Raspberry Pi

- Run the following two commands in a terminal on the Raspberry Pi:
```
sudo apt update
sudo apt full-upgrade
```
- The following command installs Raspberry Pi Connect:
 `sudo apt install rpi-connect`

- After the installation is done run the following command to turn it on:
`rpi-connect on`

## Step 4: Link the Raspberry Pi device to a connect account

- The previous command should have opened a browser prompting you to sign in with a Raspberry Pi ID, if not do the following:

   - In the top right corner, an icon with a circle with two squares should have appeared, clicking on that and choosing “sign in” should open the browser.

- Sign in with a Raspberry Pi ID on this site, if you don’t have one it is free to create one.

- After that assign a name to the device and press the blue button.

- To be able to remote access your Raspberry Pi even after a restart the following command is required in the terminal:
`loginctl enable-linger`

Now you can access your Raspberry Pi remotely even when being on different internet if the Raspberry Pi has internet connection. You can either access it via screensharing or with a shell on the following link:

https://connect.raspberrypi.com/

This allows you to use your Raspberry Pi without needing to have a screen, keyboard or mouse connected to it, all you need is power supplied to it and internet connection.

## Installing deCONZ

This application is required for handling the communication between the ZigbeeHandler
and the Zigbee devices.

- Open the terminal and do the following commands, this installs the dependencies:
```
sudo apt update
sudo apt install i2c-tools build-essential raspberrypi-kernel-headers
```

- After that download the installation archive:
```
curl -O -L https://github.com/dresden-elektronik/raspbee2-rtc/archive/master.zip
unzip master.zip
```

- Then go into the extracted directory:
`cd raspbee2-rtc-master`

- Compile the RTC kernel module:
`make`

- Install the RTC kernel module:
`sudo make install`

- Reboot the Raspberry Pi:
`sudo reboot`

- After the Raspberry Pi has restarted the user access rights of the serial interface has to be configured, this is done by running the following command:
`sudo raspi-config`

- Set the following configurations:
```
Interface Options -> Serial Port
- Would you like a login shell accessible over serial? -> No
- Would you like the serial port hardware to be enabled? -> Yes
```

- Reboot the Raspberry Pi again for the access rights becomes active.

- When the Raspberry Pi has restarted import the Phoscon public key with the following command:
```
wget -qO- https://phoscon.de/apt/deconz.pub.key | gpg --dearmor | \
sudo tee /etc/apt/trusted.gpg.d/deconz-keyring.gpg >/dev/null
```

- After this configure the APT repository for deCONZ:
```
sudo sh -c "echo 'deb http://phoscon.de/apt/deconz \
generic main' > /etc/apt/sources.list.d/deconz.list"
```

- Then update the APT package list:
`sudo apt update`

- Lastly install deCONZ with the following command:
`sudo apt install deconz`

- The deCONZ program can now be found via the application menu:
Menu -> Programming -> deCONZ



## Configuring deCONZ and connecting Zigbee devices

**<H1>How to get an API key</H1>**
- [ ] Start **deConz** application
![deConz](https://github.com/user-attachments/assets/302f94ed-15ba-40c3-9acb-ea446fbd9fc4)
- [ ] Click the **Phoscon app** button (top right) _NOTE: This will open a browser guiding you to the phoscon page to add new devices_
![Phoscon app](https://github.com/user-attachments/assets/865517ba-0ad8-46d7-a822-4e3053a8ea62)
  - [ ] If you've already set up the gateway, click your gateway and skip to next step (Open menu/Select **Gateway**)
  - [ ] Click *Phoscon-GW*
  ![Phoscon-GW](https://github.com/user-attachments/assets/ee226bd3-7eed-43eb-aded-2d051f37aca1)
  - [ ] Fill out name, and password
    - [ ] Seems like a button is missing, click ENTER when you've typed in password again
  - [ ] Create the first group, e.g. kitchen
- [ ]  Open menu (click the three lines on the top-left part of the webpage)
![Open menu](https://github.com/user-attachments/assets/8cec7c7d-d82a-4d0e-950f-dffebf99d4d6)
- [ ] Select **Gateway**
![Gateway](https://github.com/user-attachments/assets/3eaa1ff4-cd37-4f91-8adb-7252fd25038c)
- [ ] Click **Advanced**
![Gateway Advanced](https://github.com/user-attachments/assets/476eaa3f-b14f-46da-85ff-8bd2c3ceafc1)
- [ ] Scroll down to "Authenticate app" and click it
![Authenticate](https://github.com/user-attachments/assets/280100bc-f54d-4837-a901-bcda75f21900)
- [ ] Open a command prompt and use the below command to generate an API key
- [ ] curl -v -X POST "http://localhost:8080/api" -d '{"devicetype": "ZigBee system"}'
  - [ ] Your api key is in the red box after "username":
  - [ ] e.g. "username":"A123BCD" would mean your api key is A123BCD
  ![API key generated](https://github.com/user-attachments/assets/40d0cfd2-dd1b-484b-85d4-e5b23c4993f9)

**<H1>How to install new smart thermostat</H1>**
- [ ] Start **deConz** application
![deConz](https://github.com/user-attachments/assets/302f94ed-15ba-40c3-9acb-ea446fbd9fc4)
- [ ] Click the **Phoscon app** button (top right) _NOTE: This will open a browser guiding you to the phoscon page to add new devices_
![Phoscon app](https://github.com/user-attachments/assets/865517ba-0ad8-46d7-a822-4e3053a8ea62)
  - [ ] If you've already set up the gateway, click your gateway and skip to next step (Open menu/Select **Thermostats**)
  - [ ] Click *Phoscon-GW*
  ![Phoscon-GW](https://github.com/user-attachments/assets/ee226bd3-7eed-43eb-aded-2d051f37aca1)
  - [ ] Fill out name, and password
    - [ ] Seems like a button is missing, click ENTER when you've typed in password again
  - [ ] Create the first group, e.g. kitchen
- [ ]  Open menu (click the three lines on the top-left part of the webpage)
![Open menu](https://github.com/user-attachments/assets/8cec7c7d-d82a-4d0e-950f-dffebf99d4d6)
- [ ] Select **Thermostats** in the menu
![Select thermostats](https://github.com/user-attachments/assets/cb44815e-a093-44e6-927f-cc330ef5b6b7)
  - [ ] Click **Connect new thermostat**
    - [ ] Start smart thermostat
    - [ ] Put thermostat in pairing mode (Should be found in user manual)
      - [ ] e.g. hold button 10sec to enter pairing mode
- [ ] Add a new unitasset for the smart thermostat in the systemconfig.json (There's a template for them created on system start)
  - [ ] Add a name
  - [ ] Add a model
    - [ ] "type": "ZHAThermostat" as shown in example below
  - [ ] Add the API key
  - [ ] Add uniqueid
  - [ ] Currently have to open a command prompt and use the below command to find uniqueid. Insert API key generated in [apikey]
    - [ ] ~curl -v "http://localhost:8080/api/[apikey]/sensors" | jq
    - [ ] Find the last connected device with type: "ZHAThermostat"
    Example:
![ThermostatUniqueID](https://github.com/user-attachments/assets/42a70343-044e-4b7f-bd2f-4053e6ad397a)

**<H1>How to install new power plug</H1>**
- [ ] Start **deConz** application
![deConz](https://github.com/user-attachments/assets/302f94ed-15ba-40c3-9acb-ea446fbd9fc4)
- [ ] Click the **Phoscon app** button (top right) _NOTE: This will open a browser guiding you to the phoscon page to add new devices_
![Phoscon app](https://github.com/user-attachments/assets/865517ba-0ad8-46d7-a822-4e3053a8ea62)
  - [ ] If you've already set up the gateway, click your gateway and skip to next step (Open menu/Select **Plugs**)
  - [ ] Click *Phoscon-GW*
  ![Phoscon-GW](https://github.com/user-attachments/assets/ee226bd3-7eed-43eb-aded-2d051f37aca1)
  - [ ] Fill out name, and password
    - [ ] Seems like a button is missing, click ENTER when you've typed in password again
  - [ ] Create the first group, e.g. kitchen
- [ ]  Open menu (click the three lines on the top-left part of the webpage)
![Open menu](https://github.com/user-attachments/assets/8cec7c7d-d82a-4d0e-950f-dffebf99d4d6)
- [ ] Select **Plugs** in the menu
![Select plugs](https://github.com/user-attachments/assets/2b351894-0963-42fe-9bc3-02c09dbbbef7)
  - [ ] Connect plug to outlet and turn it on
    - [ ] Put smart plug in pairing mode (Should be found in user manual)
      - [ ] e.g. hold button for 4 seconds to enter pairing mode
  - [ ] Click **Add new Plug**
    - [ ] It should have automatically found the smart plug
- [ ] Add a new unitasset for the smart thermostat in the systemconfig.json (There's a template for them created on system start)
  - [ ] Add a name
  - [ ] Add a model
    - [ ] "type": "Smart plug" as shown in example below
  - [ ] Add the API key
  - [ ] Add uniqueid
  - [ ] Add period (in seconds, used by a function to check room temp)
    - _NOTE: If the plug should be controlled by a switch, set period to 0_
  - [ ] Currently have to open a command prompt and use the below command to get uniqueid and model.
    - [ ] curl -v "http://localhost:8080/api/B3AFB6415A/lights" | jq
    - [ ] Find the last connected device with type: "ZHAPlug"
    Example:
![PlugUniqueID](https://github.com/user-attachments/assets/0b558786-075e-4c4a-b02f-c557003372a2)


**<H1>How to connect breadboard (with temperature sensor ds18b20) to Raspberry Pi</H1>**
- [ ] System configurations for Raspberry Pi
  - [ ] Enable 1-wire in raspi-config tool
      - [ ] sudo raspi-config
        - [ ] Select Advanced Option -> 1-Wire -> Yes
  - [ ] As root append "dtoverlay=w1-gpio-pullup,gpiopin=26" to /boot/firmware/config.txt with below command
    - [ ] echo "dtoverlay=w1-gpio-pullup,gpiopin=26" >> /boot/firmware/config.txt
  - [ ] Restart the Raspberry Pi to enable the settings
Above guide to enable 1-Wire was copied from [waveshare.com](https://www.waveshare.com/wiki/Raspberry_Pi_Tutorial_Series:_1-Wire_DS18B20_Sensor)
- [ ] Set up breadboard, connect temp. sensor, resistor (4k Ohm) and wires
![Image](https://github.com/user-attachments/assets/90c6ee19-a040-4fa1-b365-8116f1e1ec1c)
![Breadboard](https://github.com/user-attachments/assets/f525e9f5-2963-4a48-a6da-b9042ce5477b)
- [ ] Connect breadboard to Raspberry Pi
  - [ ] Green wire to 3.3v
  - [ ] Blue wire to ground
  - [ ] Yellow wire to GPIO 26
![Raspberry Pi](https://github.com/user-attachments/assets/003e185a-087a-40e9-87aa-e045350c1ab3)
![GPIO](https://www.etechnophiles.com/wp-content/uploads/2021/01/R-Pi-4-GPIO-Pinout-1.jpg)
Above picture copied from [etechnophiles.com](https://www.etechnophiles.com/wp-content/uploads/2021/01/R-Pi-4-GPIO-Pinout-1.jpg)
  - [ ] Completed setup
![Raspberry Pi and Breadboard](https://github.com/user-attachments/assets/a7666440-d65d-4bb9-80b5-db4fd65f19b7)


**<H1>How to install new Wireless Remote Switch</H1>**
- [ ] Start **deConz** application
![deConz](https://github.com/user-attachments/assets/302f94ed-15ba-40c3-9acb-ea446fbd9fc4)
- [ ] Click the **Phoscon app** button (top right) _NOTE: This will open a browser guiding you to the phoscon page to add new devices_
![Phoscon app](https://github.com/user-attachments/assets/865517ba-0ad8-46d7-a822-4e3053a8ea62)
  - [ ] If you've already set up the gateway, click your gateway and skip to next step (Open menu/Select **Switches**)
  - [ ] Click *Phoscon-GW*
  ![Phoscon-GW](https://github.com/user-attachments/assets/ee226bd3-7eed-43eb-aded-2d051f37aca1)
  - [ ] Fill out name, and password
    - [ ] Seems like a button is missing, click ENTER when you've typed in password again
  - [ ] Optional: Create the first group, e.g. Living room, can close the popup by pressing the X (top-right of popup)
- [ ]  Open menu (click the three lines on the top-left part of the webpage)
![Open menu](https://github.com/user-attachments/assets/8cec7c7d-d82a-4d0e-950f-dffebf99d4d6)
- [ ] Select **Switches** in the menu
![SwitchesMenu](https://github.com/user-attachments/assets/7419414e-3968-4972-ba3e-36b076c03057)
  - [ ] Click **Add new switch**
    - [ ] Start wireless remote switch
    - [ ] Put the wireless remote switch in binding/pairing mode (Should be found in user manual)
      - [ ] e.g. hold button 10sec to enter binding mode
- [ ] Add a new unitasset for the smart switch in the systemconfig.json (There's a template for them created on system start)
  - [ ] Add a name
  - [ ] Add a model
    - [ ] "type": "ZHASwitch" as shown in example below
  - [ ] Add the API key
  - [ ] Add uniqueid
  - [ ] Add slaves uniqueid (The smart power plugs/smart lights its supposed to control)
    - [ ] All lights/plugs be found by using the `curl -v "http://localhost:8080/api/[apikey]/lights" | jq` command
    Example of complete switch unit asset:
![SwitchUnitAsset](https://github.com/user-attachments/assets/1d954ca3-28ed-4113-8e50-f274d17da521)
  - [ ] Currently have to open a command prompt and use the below command to find uniqueid. Insert API key generated in [apikey]
    - [ ] `curl -v "http://localhost:8080/api/[apikey]/sensors" | jq`
    - [ ] Find the last connected device with type: "ZHASwitch"
    Example:
![SwitchUniqueID](https://github.com/user-attachments/assets/7d19bf99-703e-46a4-bf0b-76cafa3a95dc)



# Running the systems

The easiest way to run the systems in a local cloud is by using Docker and its tools.
These can first be installed by running:

```
# Add Docker's official GPG key:
sudo curl -fsSL https://download.docker.com/linux/ubuntu/gpg -o /etc/apt/keyrings/docker.asc
sudo chmod a+r /etc/apt/keyrings/docker.asc

# Add the repository to Apt sources:
echo \
  "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.asc] https://download.docker.com/linux/ubuntu \
  $(. /etc/os-release && echo "$VERSION_CODENAME") stable" | \
  sudo tee /etc/apt/sources.list.d/docker.list > /dev/null
sudo apt update
sudo apt install docker docker-compose
```

Source: https://docs.docker.com/engine/install/ubuntu/

## Prepare the server environment

Run:
```
cd /srv
git clone "https://github.com/lmas/d0020e_code.git" src
git clone "https://github.com/lmas/mbaigo-systems.git" mbaigo
```

Now `/srv/src` will contain the source code of our custom systems, while `/srv/mbaigo` will contain the source code of the core systems. Inside both directories exists `Dockerfile` that is used for building docker container images.

Next, create a symbolic link to the `docker-compose` file while still being in the `/srv/` dir:
```
ln -s src/docker-compose.yml docker-compose.ym
```

This compose file will build local container images for all systems and allow running them as containers.

## Running docker, and other docker commands

Starting the containers:
```
sudo docker-compose up -d
```

Stopping the containers:
```
sudo docker-compose down
```

Printing and tailing the container logs:
```
sudo docker-compose logs -f
```

Inspecting a running container:
```
sudo docker exec -it <mycontainer> sh
```

Rebuilding all containers:
```
sudo docker-compose build --no-cache
```
