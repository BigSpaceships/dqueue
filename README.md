## Dairyqueen / Discussion Queue / Circlejerk
Whatever you want to call this is a discussion queue used for big discussions like elections

Access it at [dq.csh.rit.edu](https://dq.csh.rit.edu) or dev at [dairyqueen.cs.house](https://dairyqueen.cs.house)

### How it works

#### Tech stack
Raw Go `net/http` package
Custom OIDC auth to work with csh and google auth together 
Static frontend, using a rest api (not templates)
Webserver to handle a lot of the connection stuff and realtime discussion

#### Code structure
Right now frontend wise everything is tossed into `index.html` and `index.js`, but this should probably be changed, at least for the javascript
Everything is non-persistent because it doesn't need to be. There is both a discussion struct and queue struct on the go side, with the discussion holing a root queue as well as map of queue to ids to enable easy accesssing of them. Each queue also stores a list of it's children and a reference to it's parent, mostly for frontend stuff.
Eventually there will be support for multiple discussions which is why that data structure exists at all but right now only one is created.

Actions performed by users are send to the server with HTTP, which then relays the data over the websocket to every client. The queue points are just sent with updates (creation and deletion) which means they could potentially get out of sync when there's a bug so both refreshing and the refresh button will reset the points. Also, when you move to a new queue it loads that using HTTP because I couldn't be bothered to put that in the websocket because that just... doesn't make sense to me.

#### Frustrations 
- css
- like seriously css

I've been writing the bootstrap html stuff before I implement everything which works GREAT except for the fact that it's finicky and getting everything to work, especially on mobile and desktop without the font being massive is a BIG headache so I'm so sick and tired of it by the time I get to do the actually fun stuff

- Also sometimes the websocket refuses to load for like 30 seconds on localhost but I think I just had like 4 tabs of dairyqueen open so it was lagging my browser with websockets, this should be investigated further because that sounds insane but idk

### Planned features
- [x] Fix mobile
- [ ] Blacklist
- [ ] Log of administrative actions
- [x] Nested queues (tree traversal fr)
- [ ] Notes for what your points were (client side)
- [ ] Improve check button UX
