# JS.BET

## Purpose
Js.bet is a online game in which you attempt to gain the greatest amount of virtual currency through betting on one of two sides of a duel.

The combatants in this duel happen to be different **Javascript Frameworks**, and there are a lot of them who wish to fight each other for supremacy.

As a fledgeling survivor of the web apocalypse, you have fealty to one or many of these frameworks, so you can bet on your framework of choice to show your faith in their efficiency, ergonomics, or whatever metric you value most.

If you gain enough of this currency, you can double or triple down on your bets to gain even more of it!

Some more cowardly survivors may hedge their bets by betting on both sides, ensuring some meager gains while demonstrating disloyalty to your patron framework.

Either way, play and enjoy the automated battles and quirky abilities!

## How to play
Simply join at [js.bet](https://js.bet) to see the current battle in progress and create an account to start placing bets.

You can choose your bet amount underneath your chosen fighter and wait until the battle concludes to receive your reward.

## About this project
This project was created initially to test out using a hypermedia approach to a multiplayer game and ended up using an architecture popularized by the [Datastar](https://data-star.dev) authors of streaming html responses to the user as new changes occur to the game state.

More specifically, the server response with [Server Sent Events](https://developer.mozilla.org/en-US/docs/Web/API/Server-sent_events/Using_server-sent_events) rather than html responses which allows for a single connection to stream out many responses to the client.

With this setup, the game state only needs to be rendered one time and simply replicated to all clients rather than performing a new render per user and fits the lock-step progression of the game much better than other approaches.

The other approaches considered include:

1. Polling: Send out the rendered page to each user once per second and assume the polling client will catch up over time.

 - Benefits: Simplified model of request/response and the server can avoid holding state for each client 
 - Downsides: Users may receive late updates or not receive important updates that happen due to inconsistent timings.

2. Websockets: Connect clients and servers with a WS connection per client, sending user interactions and server state updates via the WS connection.

 - Benefits: Opens the possibility for more rich user interactions
 - Downsides: (MAJOR) Increased server cost and complexity in managing WS connections. 

Once SSE was considered, it was quickly recognized as a far better approach for this, since the server is authoritative and clients send only a few requests to place bets. 

