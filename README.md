## Deus Ex Machina ![build status](https://circleci.com/gh/BadLamb/dexm.png?circle-token=897cb050f72c9b0b2833be16146c447fde345617)  [![Join the chat at freenode:dexm](https://img.shields.io/badge/irc-freenode:%20%23dexm-blue.svg)](http://webchat.freenode.net/?channels=%23dexm)
#### The Blockchain datacenter
DexM is a cryptocurrency that allows people to host static and dynamic websites in a distributed and censorship free way on the public
internet.

It's a PoW blockchain that allows publishing various types of contracts:
- **Smart contracts**   
These need no introduction, but unlike ETH these contracts are programmed in Javascript
If you want more flexibility you can also code your contracts in Javascript, allowing you to write them in a lot of languages that compile to Javascript(C, Typescript, Ruby, PHP, Python, Golang, Java, Dart etc).
- **Function contracts**   
Function contracts are a form of contract that responds to events like HTTP request. These contracts can be built in many
different programming languages with very little difference from building a standard serverless app, but with the advantage 
of having your site safe from takedowns and downtime. Your app will run in a Javascript vm and will use some custom apis.
- **CDN contracts**   
CDN contracts are a form of contract that is used to host static content(webpages, photos, videos) and are cached by many edge nodes around the globe. Another distinguishing feature is the regional autoscaling: this optimizes the response time of your users by caching your content closer to your users. The cost of these contracts depends on the redundancy level of your data, and the GBs of data served to clients.

But it also has some other features that will make life easier for miners and holders of the currency:
- A simple deamon that allows accepting DeXm payments with just 1 http request     
- A hotpatching mechanism that allows quickly patching clients in case of a security problem, it's also secure because multiple people are required to approve
 the patch before it gets accepted by the clients.
 
### Planned features:
- **Distributed exchange run on the network that would allow varying rewards based on the price of the coin to keep profitability of mining stable**
- **Schelling sidechain to allow decentralized voting on the blockchain.**
