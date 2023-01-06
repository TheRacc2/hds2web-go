const port = 9999; // check settings.toml

let socket = new ReconnectingWebSocket('ws://127.0.0.1:' + port + '/web');

socket.onopen = () => {
    console.log('Successfully Connected');
};

socket.onclose = event => {
    console.log('Socket Closed Connection: ', event);
};

socket.onerror = error => {
    console.log('Socket Error: ', error);
};

let animation = new CountUp('bpm', 0, 0, 0, .5, {
	useEasing: true,
    useGrouping: true,
    separator: " ",
    decimal: "."
});

function parseElement(data, name) {
    if (!data.includes(name))
        return null;

    return data.split(':')[1];
}

socket.onmessage = event => {
    let data = JSON.parse(event.data)['data']; // get data from message. This includes things like heartrate, calories, etc
    console.log(data)

    let value = parseElement(data, 'heartRate'); // will be null if something else was sent, like "calories"

    if (value !== null) {
        animation.update(value);
    }
}