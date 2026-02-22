async function fetchAPI(input, init) {
  const response = await fetch(input, init);

  if (response.status == 403) {
    console.error("Not logged in");
    window.location.assign("/auth");
    return;
  }

  return response;
}

async function getDiscussion() {
  try {
    const response = await fetchAPI(window.location.origin + "/api/discussion")

    if (!response.ok) {
      throw new Error(response.status);
    }

    const json = await response.json()

    return json
  } catch (e) {
    throw new Error(e)
  }
}

async function getQueue(id) {
  try {
    const response = await fetchAPI(`${window.location.origin}/api/queue/${id}`)

    if (!response.ok) {
      throw new Error(response.status);
    }

    const json = await response.json()

    return json
  } catch (e) {
    throw new Error(e)
  }
}

async function getPath(id) {
  try {
    const response = await fetchAPI(`${window.location.origin}/api/queue/${id}/path`)

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
  fetchAPI(`${window.location.origin}/api/queue/${window.queueId}/${type}`, {
    method: "POST",
  })
}

function leaveQueue(type, id) {
  fetchAPI(`${window.location.origin}/api/queue/${window.queueId}/${type}/${id}`, {
    method: "DELETE",
  });
}

function changeTopic(event) {
  fetchAPI(`${window.location.origin}/api/queue/${window.queueId}`, {
    method: "PATCH",
    body: JSON.stringify({
      "new-topic": event.target.value
    }),
  })
}

async function newQueue(topic) {
  if (!window.userInfo.isEboard) {
    console.error("Only E-Board members can create new queues");
    return;
  }

  fetchAPI(`${window.location.origin}/api/queue/${window.queueId}/new-child`, {
    method: "POST",
    body: JSON.stringify({
      "topic": topic,
      "move-users": true
    }),
  })
}

function newQueueFormSubmit() {
  const topicElement = document.getElementById('newQueueTopic')
  newQueue(topicElement.value)
  topicElement.value = "";
}

function joinWebsocket(retryCount = 0) {
  const isSecure = window.location.protocol == "https:";
  const protocol = isSecure ? "wss" : "ws";
  const socket = new WebSocket(`${protocol}://${window.location.host}/api/joinws`);

  socket.addEventListener("message", (event) => {
    const eventData = JSON.parse(event.data);

    switch (eventData.type) {
      case "point":
        if (eventData.queueId == window.queueId) {
          addEntryToQueue("point", eventData.data);
        }
        break;
      case "clarifier":
        if (eventData.queueId == window.queueId) {
          addEntryToQueue("clarifier", eventData.data);
        }
        break;
      case "delete":
        if (eventData.queueId == window.queueId) {
          removeEntryFromQueue(eventData.id, eventData.dismisser);
        }
        break;
      case "topic":
        if (eventData.queueId == window.queueId) {
          setTopic(eventData.topic);
        }
        break;
      case "new-queue":
        loadQueueDom(eventData.queue);
        break;
    }
  });

  socket.addEventListener("open", async () => {
    if (retryCount != 0) {
      console.log("reestablished websocket connection");
    }

    await rebuildQueue();
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
    document.getElementById("createQueueModalBtn").classList.add("d-none");
  }
}

function getListNodeForQueueEntry(queueEntry) {
  const type = queueEntry["type"];
  const name = `${queueEntry["name"]} (${queueEntry["username"]})`
  const id = queueEntry["id"];

  const listElement = document.createElement("div");
  listElement.classList.add("d-flex", "flex-row", "px-4", "py-3", "py-lg-2", "border-bottom", "align-items-center");

  listElement.dataset.id = id;

  const badgeElement = document.createElement("span");
  badgeElement.classList.add("badge", "align-self-center", "me-2")

  if (type == "point") {
    badgeElement.classList.add("text-bg-info");
    badgeElement.appendChild(document.createTextNode("Point"));
  } else if (type == "clarifier") {
    badgeElement.classList.add("text-bg-success");
    badgeElement.appendChild(document.createTextNode("Clarifier"));
  }

  listElement.appendChild(badgeElement);
  listElement.appendChild(document.createTextNode(name));

  if (queueEntry["username"] == window.userInfo.preferred_username || window.userInfo.is_eboard) {
    const completeLink = document.createElement("a");
    completeLink.href = "#";
    completeLink.classList.add("ms-auto");
    completeLink.onclick = () => { leaveQueue(type, id); return false; };
    completeLink.innerHTML = '<i class="bi bi-check-lg text-success fs-4"></i>'
    listElement.appendChild(completeLink)
  }
  return listElement;
}

function getListNodeForQueueChild(queue) {
  const topic = queue["topic"];
  const id = queue["id"];

  const listElement = document.createElement("a");
  listElement.classList.add("queue-child", "d-flex", "flex-row", "px-4", "py-3", "py-lg-2", "border-bottom", "align-items-center", "text-decoration-none");

  listElement.onclick = async () => {
    const newQueue = await getQueue(id);
    loadQueueDom(newQueue);
  };

  const badgeElement = document.createElement("span");
  badgeElement.classList.add("badge", "text-bg-secondary", "align-self-center", "me-2")
  badgeElement.appendChild(document.createTextNode("Point"));
  listElement.appendChild(badgeElement);
  listElement.appendChild(document.createTextNode(topic));

  return listElement;
}

function addChildToQueue(queue) {
  const listElement = document.querySelector("#list-parent");

  const divider = document.querySelector("div.children-spacer");
  listElement.insertBefore(getListNodeForQueueChild(queue), divider);
}

function addEntryToQueue(type, queueEntry) {
  const listElement = document.querySelector("#list-parent");

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
  const listElement = document.querySelector("#list-parent");
  Array.from(listElement.children).filter((el) => el.dataset.id == id).forEach((el) => {
    el.remove();
  })

  console.log(`${dismisser} dismissed point ${id}`);

  if (listElement.children.length == 3) {
    const emptyQueueMsg = document.querySelector("p#empty-queue")
    emptyQueueMsg.removeAttribute("hidden")
  }
}

function setTopic(topic) {
  console.log("new topic", topic)
  document.querySelector("input.discussion-title").value = topic;
}

function setPath(path) {
  const breadcrumbParent = document.querySelector("ol.breadcrumb");

  for (let i = breadcrumbParent.childElementCount - 1; i >= 0; i--) {
    breadcrumbParent.children[i].remove();
  }

  for (let i = 0; i < path.length; i++) {
    const pathEl = path[i];

    const liElement = document.createElement("li");
    liElement.classList.add("breadcrumb-item");

    const aElement = document.createElement("a");
    aElement.classList.add("link-body-emphasis", "text-decoration-none");

    if (i == path.length - 1) {
      aElement.classList.add("text-primary");
    } else {
      aElement.classList.add("text-muted");
    }

    aElement.href = "#";

    aElement.onclick = async () => {
      const newQueue = await getQueue(pathEl.id);
      loadQueueDom(newQueue);
    }

    liElement.appendChild(aElement);

    const spanElement = document.createElement("span");
    spanElement.appendChild(document.createTextNode(pathEl.topic));

    aElement.appendChild(spanElement);

    breadcrumbParent.appendChild(liElement);
  };
}

async function loadQueueDom(queue) {
  window.queueId = queue.id;

  const listElement = document.querySelector("#list-parent");

  Array.from(listElement.children).filter((el) => {
    return !el.classList.contains("clarifier-spacer") && !el.classList.contains("children-spacer") && el.id != "empty-queue"
  }).forEach((el) => {
    el.remove();
  });

  queue.children.forEach((child) => {
    addChildToQueue(child);
  });

  queue.clarifiers.forEach((queueEntry) => {
    addEntryToQueue('clarifier', queueEntry);
  });

  queue.points.forEach((queueEntry) => {
    addEntryToQueue('point', queueEntry);
  });

  setTopic(queue.topic);

  const path = await getPath(queue.id);
  setPath(path);
}

async function rebuildQueue() {
  if (window.queueId) {
    let queue = await getQueue(window.queueId);

    await loadQueueDom(queue)
  } else {
    let discussion = await getDiscussion();
    let queue = discussion.queue;

    await loadQueueDom(queue)
  }
}

async function main() {
  try {
    userInfo = await getUserInfo();
    console.log(userInfo)
    updateUserInfo(userInfo);
  } catch (e) {
  }

  joinWebsocket();

  await rebuildQueue();
}

main()
