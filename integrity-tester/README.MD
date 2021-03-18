## Preparation
* Run locally mbelt-filecoin-streamer
* Create and fill in .env file in the mbelt-filecoin-streamer/integrity-tester directory(see example.env)
## Tests 
* Check node top block with indexator top block
* Check indexator missed blocks - cid(filecoin.blocks table) might be in blocks[] (filecoin.tipsets table) by corresponding height
* Check indexator missed tipsets
## How to run tests
* cd ./mbelt-filecoin-streamer/integrity-tester
* npm install
* npm test
