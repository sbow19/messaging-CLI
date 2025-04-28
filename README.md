# Messaging CLI: Go and TView

A CLI-based real time messaging application allowing users to sign up and chat with other users ona local network in the terminal.

Features a notification bar and a game section (currently only Snake is available).

I undertook this project to learn Golang as well as to get familiar with using websockets, SQLite, and the Tview CLI builder library in Go.

At the moment, the application functions on a Debian 11 machine and has not been tested for compatibility with other operating systems. As well, I have not implemented a mechanism to 
automatically detect the subnet that the server is located on - this will need to be configured 
manually at this point. 


## Goals
  - Create a functional Real-time messaging app in the CLI using  websockets and client-server architecture.
  - Effectively implement Goroutines and channels to manage concurrent aspects of the application. 
  - Implement CRUD functionality with some complex state management across different components of the CLI app -- using SQLite.

## Demo Video

-[Searching and adding friends]()

-[Messaging]()

-[Snake game]()


## Features

- Server-side:
  - UDP broadcast over local network for client discovery.
  - SQLite instance for storing user details, connections, and messages. 

- Client-side:
  - Sign up and login functionality.
  - User search and friend requests.
  - Instant messaging functionality between friends.
  - Real time broadcasts of active status to friend network. 
  - Notifications panel showing new messages and active friends.
  - Game section (single-player only): only Snake is available. 

## Next Steps

- **Support for different operating systems**: The application was developed on Debian 11, so it has not been tested against other operating systems.
