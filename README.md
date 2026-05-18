# multicamobserver
Go Language Monolith Multicam Observer PWA.


# MULTICAMOBSERVER: A Monolithic Go-Based PWA Multicam Monitoring System

**Author:** Gilang Teja Krishna ([github.com/gtkrshnaaa](https://www.google.com/search?q=https://github.com/gtkrshnaaa))

**Personal Site:** [gtkrshnaaa.my.id](https://gtkrshnaaa.my.id)

---

## 1. Background

High mobility and the growing trend of flexible working arrangements (remote/hybrid) require software developers and professionals to frequently shift their workspaces—whether it's a home office, an outdoor patio, or a co-working space. However, leaving a workspace equipped with valuable hardware and digital assets unattended, even momentarily, poses an inherent security risk.

Current commercial security monitoring solutions (IP Cameras / CCTV) suffer from several fundamental flaws:

1. **Privacy Concerns:** The majority force users to route their video traffic through third-party cloud servers, potentially compromising personal data privacy.
2. **Closed Ecosystems:** They require the installation of native, memory-heavy applications (bloatware) and are often locked to specific hardware manufacturers.
3. **Hardware Inefficiency:** They ignore the fact that we often have legacy devices (older laptops or spare smartphones) with perfectly functional cameras that can be repurposed.

The **MulticamObserver** project was initiated as an open-source software engineering experiment to solve these problems. This system is designed as an independent signaling hub capable of transforming any device with a modern web browser into a broadcaster node. By leveraging a lightweight Go architecture and a Progressive Web App (PWA) implementation, this system proves that data privacy, security, and computational efficiency can coexist without relying on expensive commercial infrastructure.

## 2. Vision

To build an independent, extremely lightweight, and privacy-first web-based personal security monitoring ecosystem. This system aims to provide absolute peace of mind through real-time visual surveillance, without surrendering data ownership to third-party entities.

## 3. Mission

* **Extreme Computational Efficiency:** Optimizing CPU and memory usage so the application can run stably in the background on entry-level hardware. The system is specifically engineered and tested to run seamlessly even on a Linux Ubuntu machine powered by a Celeron processor and 8 GB of RAM.
* **PWA Accessibility:** Delivering a responsive, cross-platform interface. Whether accessed via a desktop monitor or a smartphone, the PWA ensures a native-like application experience without the friction of app store installations.
* **Hardware Decentralization:** Enabling logical e-cycling by turning obsolete computing devices into valuable security monitoring nodes.

---

## 4. Entity Architecture and Access Control

The application implements a linear Role-Based Access Control (RBAC) to strictly isolate data transmission and reception logic.

### A. Administrator (Viewer Node)

This entity holds absolute control over the central monitoring dashboard.

* **Security:** Protected by a strongly hashed password using the `bcrypt` algorithm. Sessions are managed statelessly using JSON Web Tokens (JWT) to minimize memory load on the server.
* **Authority:** Can view real-time metrics of online devices, render multi-camera video feeds simultaneously in a grid layout, and forcefully terminate data streams.

### B. Broadcaster (Camera Node)

The entity representing the physical device at the monitoring location (e.g., "Workspace Camera", "Front Door Camera").

* **Security:** Utilizes separate, location-specific credentials to prevent privilege escalation if a single node device is compromised.
* **Authority:** Exclusively permitted to access the media transmission page. It executes `navigator.mediaDevices` to access the hardware camera, establishes a socket connection, and continuously pushes data frames to the server. It has zero capability to view streams or access the admin dashboard.

---

## 5. System Flow

### Broadcaster Initialization (Video Sender)

1. The device at the target location accesses the application domain and logs in using broadcaster credentials.
2. The client application (Vanilla JS) requests *UserMedia* permissions from the browser.
3. Upon approval, a handshake is established with the Go server. The video signal is encoded and continuously pushed to the server.
4. The interface automatically engages a screensaver mode (dimmed or black screen) to conserve power and prevent screen burn-in during prolonged operation.

### Remote Monitoring (Video Receiver)

1. The Administrator opens the MulticamObserver PWA from a primary mobile device or external PC.
2. Following a successful JWT authentication, the main dashboard calls an API endpoint to fetch the current status of all registered nodes within the Go server's memory map.
3. The Administrator selects a specific node. A connection to the server's multiplexer is established, and the media stream is routed directly back to the Administrator's browser with minimal latency.

---

## 6. Application Layer Structure

The application is built on a "Clean Monolith" principle, avoiding heavy frontend frameworks to maintain peak performance.

* **Presentation Layer (Frontend):**
Constructed using pure HTML5, CSS3, and Vanilla JavaScript (ES6). All pages are Server-Side Rendered (SSR) utilizing Go's native `html/template` package. The inclusion of a PWA `manifest.json` and a Service Worker enables aggressive caching mechanisms for instant static asset loading.
* **Transport Layer (Media Communication):**
Initial iterations will utilize synchronized *WebSocket* (WSS) protocols for persistent data frame transmission. This architecture is prepared for a seamless future migration to a full Pion WebRTC implementation (P2P via local TURN/STUN servers).
* **Business & Server Layer (Backend):**
Relies entirely on the Go Standard Library. Go's native capability in handling I/O multiplexing and *Goroutines* ensures that every incoming camera connection is processed in an isolated, independent thread. Consequently, computational spikes in one camera node will not degrade the stability of the others.

---

## 7. Key Feature Highlights

1. **Zero-Configuration Client:** No complex setup required. Simply open the link in a browser, grant camera permissions, and the device instantly becomes an active CCTV node.
2. **Graceful Connection Handling:** Features a client-side auto-reconnect mechanism using exponential backoff. If the external Wi-Fi connection experiences temporary disruption, the node will continuously attempt to restore transmission without manual intervention.
3. **Encrypted Transport Security:** Rejects plain HTTP transmission. All data sockets and access interfaces are locked behind Transport Layer Security (HTTPS/WSS) proxies.
4. **High-Concurrency Memory Isolation:** Incoming connection allocations are managed within an In-Memory Map protected by `sync.RWMutex`, eliminating the potential for race conditions when the administrator rapidly switches views between multiple cameras.

---

## 8. Project Summary

MulticamObserver is more than just a surveillance application; it is a comprehensive demonstration of how a system-level programming language with native concurrency (Go) can be paired with modern browser APIs (PWA, MediaDevices) to create a robust, independent, and resilient architecture. This project democratizes personal security, returns data control to the user, and maximizes the utility of the hardware already available around us.
