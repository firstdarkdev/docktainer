# Docktainer

---

A simple and lightweight Build and Webserver for our documentation platform, fdd-docs.com.

This application features a lightweight build server, that allows our docs to be built and published from GitHub, as well as a lightweight webserver to serve the documentation.

In our case, each branch is a subdomain on fdd-docs.com (for example `mmode.fdd-docs.com`).


---

### Hosting this yourself

You shouldn't. Everything this repository does, is publicly available at https://fdd-docs.com.

If you REALLY need to for one reason or another, you can chat to us in [Discord](https://discord.firstdark.dev)

---

### Tech Used

- GO - Main Web and Build Server
- Docker - For building the environment needed for the app to work, as well as running it
- [Retype](https://retype.com) - The main framework used for building the documentation from Markdown
- [Docusaurus](https://docusaurus.io/) - Used for building Docusaurus powered websites
- Cloudflare - HTTPS and other protections

---

### License

This repository and code, is licensed under the MIT license. The documentation it handles however, is licensed under ARR.