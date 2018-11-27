# coders-strike-back-referee
Brutaltester compatible referee for coders strike back

As far as I can tell this exactly matches the expected output from the server - this is tested with the validation script.

To build type:
go build csbref.go


Validation:

validate.py csbref replay.json
This will take a json replay from Codingame and given only the checkpoints / moves sent by each player checks what the referee would output against the replay. Over 500 replays have been tested without any visible errors.

This checks:
- Pod positions, speed, rounded angles and next checkpoint.
- Overall result.
- Intial pod locations (generated from checkpoints)



