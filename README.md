## Dairyqueen / Discussion Queue / Circlejerk
Whatever you want to call this is a discussion queue used for big discussions like elections

Access it at [dq.csh.rit.edu](https://dq.csh.rit.edu) or dev at [dairyqueen.cs.house](https://dairyqueen.cs.house)

### How to use it
1. Go to one of the links above
2. Log in to your provider of choice, CSH preferred
3. Click the buttons to enter yourself into the queue
   - Woah it's a Discussion Queue
4. If there are other open topics, clicking them in the list should take you to those topics until you reach where you would like to talk
5. You can click the green check to clear your point, or others points if you're E-Board
6. E-Board memebers can also change the topic, as well as create a new one which will be nested under the current discussion

### Running locally
Copy `.env.example` to `.env` and fill out with sane values, talk to and RTP or me for help with this, you may have to create a google client for this, [this guide](https://developers.google.com/identity/sign-in/web/sign-in) can help

`NON_EBOARD_ADMINS` is a comma separated list of CSH usernames that should be considered E-Board (mostly for testing)

Dairy queen can be run locally with podman/docker using the compose file using watch, which will reload frontend changes without you having to restart the server
- (I couldn't figure out how to make go work nicely with it and it's not that painful ngl)
```sh
podman compose watch
```

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
