# DynDNS - Dynamic DNS

DynDNS is a tool to update public IP address of the network the tool is running on. 
This relies on [ipify.org](https://www.ipify.org/)'s API to get the public IP of your network.

## Currently, supported DNS providers:
- Cloudflare

## Environment variables
- `CF_token`: Cloudflare API Token
- `Domain`: Domain name to use
- `Subdomain`: Subdomain of the Domain above to set the record
- `CF_proxy_disabled`: If set this will disable the proxy feature of the Cloudflare. **Enabled by default**

```bash
export CF_token=''
export Domain=''
export Subdomain=''
```
### Cloudflare API token
Follow the [Managing API Tokens and Keys](https://support.cloudflare.com/hc/en-us/articles/200167836-Managing-API-Tokens-and-Keys) documentation to create an API token for Cloudflare 
