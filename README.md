## Hybrid PoW Challenge System: Hashcash (CPU-bound) and Argon2 (Memory-bound)
### Overview
This project implements a hybrid Proof-of-Work (PoW) system that dynamically switches between CPU-bound and memory-bound challenges for client requests. The server randomly selects which type of PoW to issue:

- CPU-bound challenge (Hashcash): A lightweight challenge where the client must find a valid nonce to meet a certain difficulty level, such as finding a hash with a specific number of leading zeros.

- Memory-bound challenge (Argon2): A more resource-intensive challenge that requires the client to solve a problem using the Argon2 memory-hard hashing function, which is designed to resist hardware attacks by requiring significant memory resources.

This hybrid approach makes it harder for attackers to optimize their infrastructure for one type of challenge, adding versatility to our defense mechanisms.

### Features
- Random Challenge Selection: For each incoming request, the server randomly decides whether to issue a CPU-bound or memory-bound challenge. This randomness ensures unpredictability for clients attempting to solve challenges.

- CPU-bound (Hashcash): This challenge is lightweight, requiring the client to solve it using only CPU power, typically by finding a nonce that results in a hash with a certain number of leading zeros.

- Memory-bound (Argon2): This challenge is harder for clients to solve, as it requires a significant amount of memory and computational resources, making it ideal for preventing attacks from specialized hardware like GPUs or ASICs.

- Hybrid Security: Combining two different PoW methods makes the system robust, allowing flexibility to defend against various types of attacks (e.g., brute-force, DDoS, spam).

### Challenge Flow
1. Server-side Decision:
- Upon receiving a client request, the server decides whether to issue a CPU-bound or memory-bound challenge.
- This decision is based on a random selection process, making it unpredictable for clients or attackers.
2. CPU-bound Challenge (Hashcash):
- If a CPU-bound challenge is chosen, the client receives a challenge (a random string or nonce) and must find a value (nonce) such that the hash of the combined challenge meets the specified difficulty (e.g., a certain number of leading zeros).
3. Memory-bound Challenge (Argon2):
- If a memory-bound challenge is chosen, the client is given a challenge and must compute the Argon2 hash using the provided parameters (including memory usage, time, and parallelism). The client sends back the solution (derived key and salt) to be verified.
4. Solution Verification:
- The client submits their solution (nonce for Hashcash or derived key and salt for Argon2) back to the server.
- The server verifies the solution by recomputing the result and comparing it to the expected output.
- If the solution is correct, the client is allowed to proceed. Otherwise, the client may be asked to retry the challenge.

## Usage

```
make start
```