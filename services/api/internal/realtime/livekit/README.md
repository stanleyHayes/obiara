# LiveKit token adapter

This outbound adapter is intentionally stateless. It signs a bounded join token
and returns it directly; it never stores tokens, API secrets, participant
directories, or room listings. Consequently MongoDB and Testcontainers are not
part of this boundary. Durable room authorization remains an upstream
application concern.
