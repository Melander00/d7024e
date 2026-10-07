const express = require("express")
const bodyParser = require("body-parser")
const cors = require("cors")

const app = express()

app.use(bodyParser.json())
app.use(cors())

app.listen(3000, () => {
    console.log("Started DNS")
})


app.get("/pk/:domain", (req, res) => {
    const domain = req.params.domain

    const pk = getPk(domain)

    if(pk === undefined) {
        res.status(404).send("Domain not claimed")
        return;
    } 

    res.json({
        pk
    })
})

app.post("/claim", (req, res) => {
    const { domain, pk } = req.body;

    if(!domain || !pk) {
        res.status(400).send("Please include domain and pk in body")        
        return;
    }

    const f = setPk(domain, pk)
    if(!f) {
        res.status(401).send("Domain already claimed.")
        return;
    }

    res.send("Ok")
})

const store = new Map()

function setPk(domain, pk) {
    if(store.has(domain)) {
        return false
    }
    store.set(domain, pk)
    return true;
}

function getPk(domain) {
    return store.get(domain)
}