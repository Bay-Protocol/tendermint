const axios = require('axios')

async function main () {
  const args = process.argv.slice(2)
  const from = Number(args[0]) || 0
  const to = Number(args[1]) || 0

  for (let i = from; i < to; i++) {
    const value = Buffer.from(i.toString()).toString('base64')
    axios.get(`http://localhost:26657/broadcast_tx_commit?tx="${value}"`)
      .then(resp => console.log(resp.data))
      .catch(resp => console.error(resp.response.data))
  }
}

main()
  .catch(error => {
    console.error(error)
    process.exit(1)
  })
