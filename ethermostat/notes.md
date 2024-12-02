
# Work order

Create thing.go:

- setup the unit asset
- add required methods to meet interface
- add constructor for making default unit asset
- add constructor for creating unit asset based on config

Create main.go:

- create new system and associated husk
- create default unit asset
- try loading config for system
- load individual unit assets and aossciate them with the system
- generate certs and register system
- run web servers and wait for shutdown
- add web handlers
