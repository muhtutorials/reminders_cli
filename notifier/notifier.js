const path = require("path");

const express = require("express");
const bodyParser = require("body-parser");
const notifier = require("node-notifier");

const app = express();
const port = process.env.PORT || 5000;

app.use(bodyParser.json());

app.get("/health", (req, res) => {
    res.status(200).send();
});
app.post("/notify", (req, res) => {
    notify(req.body, reply => res.send(reply));
});

const notify = ({ title, message }, callback) => {
    notifier.notify(
        {
            title: title || "Unknown title",
            message: message || "Unknown message",
            icon: path.join(__dirname, "scorpion.jpg"),
            sound: true,
            wait: true,
            reply: true,
            closeLabel: "Completed?",
            timeout: 15
        },
        (err, response, reply) => {
            callback(reply)
        }
    );
}

app.listen(port, () => {
    console.log(`Server is up and running on port: ${port}...`);
});