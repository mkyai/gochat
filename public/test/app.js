
let currentChannel = null;
let socket = null;
let userId = null;
localStorage.clear();
function loginSignup() {
    const text = document.getElementById('loginSignup');
    if (text.textContent === 'Login') {
        document.getElementById('loginBtn').style.display = 'none';
        document.getElementById('signupBtn').style.display = 'block';
        document.getElementById('nameInput').style.display = 'block';
        text.textContent = 'Signup';
    } else {
        document.getElementById('loginBtn').style.display = 'block';
        document.getElementById('signupBtn').style.display = 'none';
        document.getElementById('nameInput').style.display = 'none';
        text.textContent = 'Login';
    }
}
// lovneet123

async function getChannels(token) {
    document.getElementById('loginBox').style.display = 'none';
    document.getElementById('channelBox').style.display = 'block';
    const localToken = localStorage.getItem('token');
    const [userResponse, channelResponse] = await Promise.all([
        fetch('http://localhost:3020/users', {
            headers: {
                Authorization: `Bearer ${token || localToken}`,
            },
        }),
        fetch('http://localhost:3020/channels', {
            headers: {
                Authorization: `Bearer ${token || localToken}`,
            },
        })
    ]);
    if (userResponse.status === 200) {
        const users = await userResponse.json();
        const userDiv = document.querySelector('.users');
        userDiv.innerHTML = '';
        users.forEach(user => {
            if (user.id === localStorage.getItem('id')) return;
            const button = document.createElement('button');
            button.textContent = `${user.name} @${user.username}`;
            button.classList.add('user-button');
            button.onclick = () => createChannel(user.id);
            userDiv.appendChild(button);
        });
    }
    if (channelResponse.status === 200) {
        const channels = await channelResponse.json();
        console.log({ channels })
        const channelListDiv = document.querySelector('.channels');
        channelListDiv.innerHTML = '';
        channels.forEach(channel => {
            const button = document.createElement('button');
            button.textContent = `${channel?.memberDetails[0]?.name} @${channel?.memberDetails[0]?.username}`;
            button.classList.add('user-button');
            button.onclick = () => connectWebSocket(channel._id, channel?.memberDetails[0]?.name, channel?.memberDetails[0]?.username);
            channelListDiv.appendChild(button);
        });
    }
}

async function createChannel(userId) {
    const token = localStorage.getItem('token');
    const response = await fetch('http://localhost:3020/channels', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
            Authorization: `Bearer ${token}`,
        },
        body: JSON.stringify({ member: userId }),
    });
    if (response.status === 201) {
        const channel = await response.json();
        console.log({ channel });
        getChannels(token);
    } else if (response.status === 200) {
        alert('Channel already exists');
    } else {
        alert('Error creating channel');
    }

}

async function login() {
    const username = document.getElementById('usernameInput').value;
    const password = document.getElementById('passwordInput').value;
    const response = await fetch('http://localhost:3020/login', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify({ username, password }),
    });
    if (response.status === 200) {
        const { token, ...data } = await response.json();
        localStorage.setItem('token', token);
        localStorage.setItem('id', data.id);
        userId = data.id;
        localStorage.setItem('user', JSON.stringify(data));
        getChannels(token);
    } else {
        alert('Invalid username or password');
    }
}


async function signup() {
    const username = document.getElementById('usernameInput').value;
    const password = document.getElementById('passwordInput').value;
    const name = document.getElementById('nameInput').value;
    const response = await fetch('http://localhost:3020/signup', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify({ username, password, name }),
    });
    if (response.status === 201) {
        const { token, ...data } = await response.json();
        localStorage.setItem('token', token);
        localStorage.setItem('id', data.id);
        userId = data.id;
        localStorage.setItem('user', JSON.stringify(data));
        getChannels(token);
    } else {
        alert('Username already exists');
    }
}

async function getMessages(channelId) {
    const token = localStorage.getItem('token');
    const response = await fetch(`http://localhost:3020/messages/${channelId}`, {
        headers: {
            Authorization: `Bearer ${token}`,
        },
    });
    if (response.status === 200) {
        const messages = await response.json();
        const messagesDiv = document.querySelector('.messages');
        messagesDiv.innerHTML = '';
        messages.forEach(message => {
            const messageDiv = document.createElement('div');
            messageDiv.textContent = message.Content;
            messageDiv.classList.add('message');
            messageDiv.style[message.CreatedBy === userId ? 'margin-left' : 'margin-right'] = 'auto';
            messagesDiv.appendChild(messageDiv);
        });
    }
}

function connectWebSocket(channelId, name, username) {
    document.getElementById('chatBox').style.display = 'block';
    document.getElementById('channelBox').style.display = 'none';
    document.getElementById('user_name').textContent = `${name}`;
    document.getElementById('user_username').title = `@${username}`;

    getMessages(channelId);

    if (socket) {
        socket.close();
    }
    token = localStorage.getItem('token');
    if (!token) {
        alert('You need to login first');
        return;
    }

    socket = new WebSocket(`ws://localhost:3020/ws/${channelId}?token=${encodeURIComponent(token)}`);

    socket.onopen = function (event) {
        console.log('Connection opened:', event);
    };

    socket.onmessage = function (event) {
        const message = JSON.parse(event.data);
        const messagesDiv = document.querySelector('.messages');
        const messageDiv = document.createElement('div');
        messageDiv.textContent = message.message;
        messageDiv.classList.add('message');
        messageDiv.style[message.userId === userId ? 'margin-left' : 'margin-right'] = 'auto';
        messagesDiv.appendChild(messageDiv);
    };

    socket.onclose = function (event) {
        console.log('Connection closed:', event);
    };

    socket.onerror = function (error) {
        console.log('WebSocket error:', error);
    };

    currentChannel = channelId;
}

function sendMessage() {
    const messageInput = document.getElementById('messageInput');
    const message = messageInput.value;

    if (socket && socket.readyState === WebSocket.OPEN) {
        socket.send(message);
        messageInput.value = '';
    }
}
