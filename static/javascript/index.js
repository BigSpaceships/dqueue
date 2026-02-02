async function getQueue() {
  try {
    const response = await fetch(window.location.origin + "/api/queue")

    if (!response.ok) {
      throw new Error(response.status);
    }

    const json = await response.json()

    return json
  } catch (e) {
    throw new Error(e)
  }
}

function enterQueue(type) {
  fetch(`${window.location.origin}/api/${type}`, {
    method: "POST",
  })
}

function leaveQueue(type, id) {
  fetch(`${window.location.origin}/api/${type}/${id}`, {
    method: "DELETE",
  });
}

function changeTopic(event) {
  fetch(`${window.location.origin}/api/change-topic`, {
    method: "POST",
    body: JSON.stringify({
      "new-topic": event.target.value
    }),
  })
}

function joinWebsocket(retryCount = 0) {
  const isSecure = window.location.protocol == "https:";
  const protocol = isSecure ? "wss" : "ws";
  const socket = new WebSocket(`${protocol}://${window.location.host}/api/join_ws`);

  socket.addEventListener("message", (event) => {
    const eventData = JSON.parse(event.data);

    switch (eventData.type) {
      case "point":
        addEntryToQueue("point", eventData.data);
        break;
      case "clarifier":
        addEntryToQueue("clarifier", eventData.data);
        break;
      case "delete":
        removeEntryFromQueue(eventData.id, eventData.dismisser);
        break;
      case "topic":
        console.log(event);
        setTopic(eventData.topic);
        break;
    }
  });

  socket.addEventListener("open", async () => {
    if (retryCount != 0) {
      console.log("reestablished websocket connection");
      await rebuildQueue();
    }
  });

  socket.addEventListener("close", (event) => {
    if (!event.wasClean) {
      setTimeout(() => joinWebsocket(retryCount + 1), 500);
    }
  });
}

function parseJwt(token) {
  if (!token) {
    return;
  }
  const base64Url = token.split('.')[1];
  const base64 = base64Url.replace(/-/g, '+').replace(/_/g, '/');

  const jsonPayload = decodeURIComponent(window.atob(base64).split('').map(function(c) {
    return '%' + ('00' + c.charCodeAt(0).toString(16)).slice(-2);
  }).join(''));

  return JSON.parse(jsonPayload);
}

async function getUserInfo() {
  const cookie = await cookieStore.get("Auth");

  const jwt = parseJwt(cookie.value)

  if (Date.now() / 1000 > jwt.exp) {
    throw new Error("Not authed");
  }

  return jwt.UserInfo;
}

function updateUserInfo(userInfo) {
  window.userInfo = userInfo;
  const name = userInfo.name;
  const profileUrl = userInfo.picture;
  const isEboard = userInfo.is_eboard;
  document.getElementById("profile-pic").src = profileUrl;
  document.getElementById("profile-name").innerText = name;

  if (!isEboard) {
    document.querySelector("input.discussion-title").setAttribute("disabled", "true");
  }
}

function getListNodeForQueueEntry(queueEntry) {
  const type = queueEntry["type"];
  const name = `${queueEntry["name"]} (${queueEntry["username"]})`
  const id = queueEntry["id"];

  const listElement = document.createElement("li");
  listElement.classList.add("list-group-item", "d-flex", "flex-row");

  listElement.dataset.id = id;

  const badgeElement = document.createElement("span");
  badgeElement.classList.add("badge", "align-self-center", "mr-1")

  if (type == "point") {
    badgeElement.classList.add("badge-info");
    badgeElement.appendChild(document.createTextNode("Point"));
  } else if (type == "clarifier") {
    badgeElement.classList.add("badge-success");
    badgeElement.appendChild(document.createTextNode("Clarifier"));
  }

  listElement.appendChild(badgeElement);
  listElement.appendChild(document.createTextNode(name));

  if (queueEntry["username"] == window.userInfo.preferred_username || window.userInfo.is_eboard) {
    const completeLink = document.createElement("a");
    completeLink.href = "#";
    completeLink.classList.add("ml-auto");
    completeLink.onclick = () => { leaveQueue(type, id); return false; };
    completeLink.innerHTML = '<i class="fa-solid fa-check text-success"></i>'
    listElement.appendChild(completeLink)
  }
  return listElement;
}

function addEntryToQueue(type, queueEntry) {
  const listElement = document.querySelector("ul.list-group");

  if (type == "clarifier") {
    const divider = document.querySelector("div.clarifier-spacer");
    listElement.insertBefore(getListNodeForQueueEntry(queueEntry), divider);
  } else if (type == "point") {
    listElement.appendChild(getListNodeForQueueEntry(queueEntry));
  } else {
    console.error("unknown type: " + type);
  }

  const emptyQueueMsg = document.querySelector("p#empty-queue")
  emptyQueueMsg.setAttribute("hidden", true)
}

function removeEntryFromQueue(id, dismisser) {
  const listElement = document.querySelector("ul.list-group");
  Array.from(listElement.children).filter((el) => el.dataset.id == id).forEach((el) => {
    el.remove();
  })

  console.log(`${dismisser} dismissed point ${id}`);

  $('.toast').toast('show')

  if (listElement.children.length == 1) {
    const emptyQueueMsg = document.querySelector("p#empty-queue")
    emptyQueueMsg.removeAttribute("hidden")
  }
}

function setTopic(topic) {
  console.log("new topic", topic)
  document.querySelector("input.discussion-title").value = topic;
}

async function rebuildQueue() {
  let queue = await getQueue();

  const listElement = document.querySelector("ul.list-group");
  Array.from(listElement.children).filter((el) => !el.classList.contains("clarifier-spacer")).forEach((el) => {
    el.remove();
  })

  queue.clarifiers.forEach((queueEntry) => {
    addEntryToQueue('clarifier', queueEntry);
  })

  queue.points.forEach((queueEntry) => {
    addEntryToQueue('point', queueEntry);
  })

  setTopic(queue.topic);
}

async function main() {
  try {
    userInfo = await getUserInfo();
    updateUserInfo(userInfo);
  } catch {
    window.location.replace(window.location.origin + "/auth")
  }

  joinWebsocket();

  $('.toast').toast({
    autohide: false
  })

  await rebuildQueue();
}

main()
