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

function joinWebsocket() {
  const isSecure = window.location.protocol == "https:";
  const protocol = isSecure ? "wss" : "ws";
  const socket = new WebSocket(`${protocol}://${window.location.host}/api/join_ws`);

  socket.addEventListener("message", (event) => {
    const eventData = JSON.parse(event.data);

    switch (eventData.type) {
      case "point":
        addEntryToQueue("point", eventData.data)
        break;
      case "clarifier":
        addEntryToQueue("clarifier", eventData.data)
        break;
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

  return parseJwt(cookie.value).UserInfo;
}

function updateUserInfo(userInfo) {
  const username = userInfo.preferred_username;
  const name = userInfo.name;
  document.getElementById("profile-pic").src = `https://profiles.csh.rit.edu/image/${username}`;
  document.getElementById("profile-name").innerText = name;
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

  const completeLink = document.createElement("a");
  completeLink.href = "#";
  completeLink.classList.add("ml-auto");
  completeLink.onclick = () => { leaveQueue(type, id); return false; };
  completeLink.innerHTML = '<i class="fa-solid fa-check text-success"></i>'

  listElement.appendChild(badgeElement);
  listElement.appendChild(document.createTextNode(name));
  listElement.appendChild(completeLink)
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
}

async function main() {
  joinWebsocket();

  userInfo = await getUserInfo();
  updateUserInfo(userInfo);

  await rebuildQueue();
}

main()
