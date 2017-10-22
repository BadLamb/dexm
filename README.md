## Deus Ex Machina ![build status](https://circleci.com/gh/BadLamb/dexm.png?circle-token=897cb050f72c9b0b2833be16146c447fde345617)[![Join the chat at freenode:dexm](https://img.shields.io/badge/irc-freenode:%20%23dexm-blue.svg)](http://webchat.freenode.net/?channels=%23dexm)
#### The Blockchain datacenter
DexM is a cryptocurrency that allows people to host static and dynamic websites in a distributed and censorship free way on the public
internet.

It's a PoW blockchain that allows publishing various types of contracts:
- **Smart contracts**   
These need no introduction, but unlike ETH these contracts are programmed in a very simple decision tree
programming language. This is for reducing the possibility of another future DAO theft, another way the DexM Smart
contracts differ from ETH is that they aren't compiled nor run in a VM, considering the very simple format of the decision
tree we can emulate them at a decent speed and not lose the original source when putting them on the blockchain.
If you want more flexibility you can also code your contracts in webassembly, allowing you to write them in (almost) any language.
- **Function contracts**   
Function contracts are a form of contract that responds to events like HTTP request. These contracts can be built in many
different programming languages with very little difference from building a standard serverless app, but with the advantage 
of having your site safe from takedowns and downtime. Your app will run in a webassembly/node vm and will use some custom apis.
- **CDN contracts**   
CDN contracts are a form of contract that is used to host static content(webpages, photos, videos) and are cached by many edge nodes around the globe. Another distinguishing feature is the regional autoscaling: this optimizes the response time of your users by caching your content closer to your users. The cost of these contracts depends on the redundancy level of your data, and the GBs of data served to clients.

But it also has some other features that will make life easier for miners and holders of the currency:
- Schelling datafeed to keep currency price stable by modifying block rewards    
- A simple deamon that allows accepting DeXm payments with just 1 http request     
- A hotpatching mechanism that allows quickly patching clients in case of a security problem, it's also secure because multiple people are required to approve
 the patch before it gets accepted by the clients.    
- A schelling mining mechanism that gives small continous rewards to miners

### Planned features:
- **Indistinguishability Obfuscation for saving secrets on the blockchain with Intel SGX support if we can get a license**
- **Distributed exchange run on the network that would allow varying rewards based on the price of the coin to keep profitability of mining stable**
