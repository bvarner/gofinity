# What is 'gofinity'
Gofinity is a library for interacting with the Carrier Infinity / Bryant Evolution communicating residential HVAC systems.

This library exports
 * Bus communication snooping / decoding
 * Bus device state tracking
 * Tools to help support reverse engineering device communications / internals

The code in this repository can loosely trace it's heritage back to 
[Andrew Danforth's Infinitive project](https://github.com/acd/infinitive) which aims to emulate a SAM device.

## History
This project has evolved out of my desire to data-log everything on my HVAC system.
I'm particularly interested in:
 * logging all temperature sensor data (my outdoor unit is always wrong when it's a sunny day) and correlating it with
   3rd party metrics / measurements (weather stations).
 * logging BTUs produced / moved by the system.
 * logging energy consumption.
 * Calculating the effective R-Value of my building envelope.
 
Longer-term, I want to be able to:
 * Control my system like a SAM. (but have it actually work with Bryant systems)
 * Segment the RS-485 network to interdict system commands to subordinate devices.
 * Create a zoning control module.
 * Build an 'economizer' module that transparently works in conjunction with the heat-pump.
  
I have a fancy hybrid heat pump & variable throttled furnace, and my utility company costs fluctuate enough that given
specific environmental factors, it may be less expensive to use gas or electricity (heat pump) for heating. I also live 
in a very temperate zone of the north american continent, where for roughly 8 months of the year the outdoor ambient air
could be used for cooling with a fan rather than running the compressor on the heat pump. 

My night-time 'set back' with the thermostat in 'auto' mode results in my heat pump compressor activating, when I could
economize with just ambient air from outside.

The HVAC company solution to outdoor air involves a heat exchanger to try to equalize incoming ambient air temperatures
to the same temperature as the conditioned space. Using this system to try and 'cool' a home would result in the fan
having to blow for considerably longer than otherwise necessary if conditioned air were simply exhausted and exchanged
for ambient outside air.

This all came about, as a consequence of talking with my HVAC tech (whom I also know outside of his professional life) 
about my economizer idea. Turns out, there are commercial products out there that do this kind of thing, but they're
pretty expensive and don't do a great job of integrating with the Carrier / Bryant communicating systems. My friend did
the opposite of trying to talk me out of it.

Challenge Accepted.


# Protocol / Device Information

## Physical ABCD Bus
**Half Duplex RS-485 + 24VAC = ABCD**

Inside Carrier Infinity / Bryant Evolution systems there's a set of screw terminals labeled, "ABCD". These are the very 
wires which connect the thermostat, air handler, and any outdoor units (heat pump, air conditioner).

The A & B terminals carry RS-485 serial data at 38400, 8n1. The C & D Terminals provide 24VAC.

## RS-485 Hardware
You'll need some way to communicate with the RS-485 bus of the ABCD connectors on your HVAC system.
Instead of dealing with the shortcomings of common USB transceivers and looking more toward my final goals of embedded 
custom hardware & software, I designed and built [pi485](https://github.com/bvarner/pi485), a TTL Serial to RS-485 
transceiver for Arduino and Raspberry Pi type devices. These are not difficult to make, and a full BOM will cost about
$15 USD (2017 prices) in single quantities. You could probably do it for less than that if you have a bunch of things in
your 'junk drawer'.

Or, you could go the RS-485 -> USB Converter route. Any of these can work.

Beware hardware that doesn't have a way to disable termination resistors, as that 
will cause issues with your ABCD bus. Also, devices which don't properly bias the AB lines with the DC supply voltage 
and 'ground', may have issues as well. These issues (and I wanted to use the UART rather than USB) were why I created 
[pi485](https://github.com/bvarner/pi485).

## Protocol Basics


### Bootup Sequence


### 


