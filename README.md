# HAR file management library

A Golang library for parsing and managing HAR (HTTP Archive) files. Provides easy-to-use structs and functions for loading, inspecting, and manipulating HAR files, making HTTP traffic analysis and debugging simpler. This library is designed to be used with the following libraries:

* [`bogdanfin/tls-client`](https://github.com/bogdanfinn/tls-client)
* The standard **`net/http`** library
* Other **custom request/response structures**

## Purpose

* Provide a **complete and persistent history of requests and responses**.
* Facilitate **monitoring, tracking, and debugging** of request systems.
* Offer a **1:1 equivalent of HAR exports** produced by tools like **Charles Proxy** or **Proxyman** during SSL proxying.

## Functional Requirements / Specifications

* **Multi-source compatibility:** Capture and generate HAR files from **tls-client**, **net/http**, or any **custom structs**.
* **Maximum fidelity to the HAR standard:** Match as closely as possible the format and content produced by **Charles Proxy** (as a reference for export quality).
* **Strict header order preservation:** Maintain the exact order of headers as defined by the TLS layer (which Go does not guarantee by default—requires specific handling).
* **Simplified API:**
  * Create a **HAR session** via a dedicated function.
  * Add a **request** to the HAR.
  * Add the **corresponding response** via a complementary function.
  * Explicit handling of **requests without responses** (timeouts, cancellations, network errors).
* **Additional fields beyond the HAR standard:**
  * **IP address** used during the TLS connection.
  * **Session ID** for session monitoring and tracking.

## Bonus / Potential Extensions

* **Monitoring integration:** Native export compatible with **Prometheus / Grafana**, or extract metrics directly from HAR files for real-time visualization.
* **Advanced historization:** Automatic HAR file storage in an **S3 bucket**, with associated **metadata/tags**:
  * Final request status: **success**, **failure**, **timeout**, **HTTP 5xx**, etc.
  * **Category / service / user tagging** for easier identification.
