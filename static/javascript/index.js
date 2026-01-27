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
  fetch(window.location.origin + "/api/enter", {
    method: "POST",
    body: JSON.stringify({
      type: type
    })
  })
}

function joinWebsocket() {
  const isSecure = window.location.protocol == "https:";
  const protocol = isSecure ? "wss" : "ws";
  const socket = new WebSocket(`${protocol}://${window.location.host}/api/join_ws`);

  socket.addEventListener("message", (event) => {
    const eventData = JSON.parse(event.data);

    switch (eventData.type) {
      case "new-point":
        addEntryToQueue(eventData.data)
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
  const name = ` ${queueEntry["name"]} (${queueEntry["username"]})`

  const listElement = document.createElement("li");
  listElement.classList.add("list-group-item");

  const badgeElement = document.createElement("span");
  badgeElement.classList.add("badge")

  if (type == "NewPoint") {
    badgeElement.classList.add("badge-info");
    badgeElement.appendChild(document.createTextNode("New Point"));
  } else if (type == "Clarifier") {
    badgeElement.classList.add("badge-success");
    badgeElement.appendChild(document.createTextNode("Clarifier"));
  }

  listElement.appendChild(badgeElement);
  listElement.appendChild(document.createTextNode(name));
  return listElement;
}

function addEntryToQueue(queueEntry) {
  const listGroupElement = document.querySelector("ul.list-group");

  listGroupElement.appendChild(getListNodeForQueueEntry(queueEntry));
}

async function main() {
  joinWebsocket()

  userInfo = await getUserInfo()
  updateUserInfo(userInfo)

  let queue = await getQueue()
  const listGroupElement = document.querySelector("ul.list-group");
  while (listGroupElement.children.length != 0) {
    listGroupElement.removeChild(listGroupElement.firstChild)
  }

  queue.forEach((queueEntry) => {
    listGroupElement.appendChild(getListNodeForQueueEntry(queueEntry));
  })
}

main()
