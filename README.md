## Deus Ex Machina
#### The Blockchain datacenter
DexM is a cryptocurrency that allows people to host static and dynamic websites in a distributed and censorship free way on the public
internet.

It's a PoW blockchain that allows publishing various types of contracts:
- **Smart contracts**   
These need no introduction, but unlike ETH these contracts are programmed in a very simple decision tree
programming language. This is for reducing the possibility of another future DAO theft, another way the DexM Smart
contracts differ from ETH is that they aren't compiled nor run in a VM, considering the very simple format of the decision
tree we can emulate them at a decent speed and not lose the original source when putting them on the blockchain.
- **Function contracts**   
Function contracts are a form of contract that responds to events like HTTP request. These contracts can be built in many
different programming languages with very little difference from building a standard serverless app, but with the advantage 
of having your site safe from takedowns and downtime. The price of execution depends on the amount of checks specified by the 
programmer, the PoW DDOS protection and the CPU time occupied from the request. By default apps run in 1 core/128MB of RAM nodes
but these can be made larger for a higher at a cost. Your app will run in a very tightly sandboxed Linux envoirement that won't allow any
syscalls but exit, rt_sigreturn, read and write on allowed files, and the IPC mechanisms needed to supply the response to the user.
- **CDN contracts**   
CDN contracts are a form of contract that is used to host static content(webpages, photos, videos) and are cached by many edge nodes around the globe.
Another distinguishing feature is the regional autoscaling: this optimizes the response time of your users by caching your content closer to 
your users. The cost of these contracts depends on the redundancy level of your data, and the GBs of data served to clients.
- **DB Contracts**   
DB contracts give you access to a public distributed KV DB. PublicDBs are a way of storing data across nodes, all the data on it is public but it isn't 
a big problem, we have developed systems that allow to publicly share many normally private information such as password hashes. We are working on a private
DB but it isn't coming anytime soon.

### Planned features:
- **Indistinguishability Obfuscation**
- **Distributed exchange run on the network that would allow varying rewards based on the price of the coin to keep profitability of mining stable**