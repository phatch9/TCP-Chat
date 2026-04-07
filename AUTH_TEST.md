# 🧪 TCP-Chat Authentication - Quick Test Guide

## Quick Start Testing

### 1. Build & Run

```bash
cd /Users/phch/Documents/TCP-Chat
go build .
./TCP-Chat
```

**Expected Output:**
```
╔══════════════════════════════════════════╗
║   TCP-Chat Server Started (v2.0 Auth)    ║
║   Listening on localhost:8888            ║
╠══════════════════════════════════════════╣
║   New users must register or login:      ║
║   /register username password           ║
║   /login username password              ║
║   Type /help for more commands           ║
╚══════════════════════════════════════════╝
```

### 2. Connect Client (in new terminal)

```bash
telnet localhost 8888
```

---

## Test Scenarios

### Scenario 1: Basic Registration & Login ✅

```bash
# Type these commands in order:

/register alice mypassword123
# Expected: ✓ Account created for 'alice'. Use /login to sign in.

/help
# Expected: Shows authentication-required help (no chat commands yet)

/login alice mypassword123
# Expected: ✓ Welcome back, alice! You are now authenticated.
#          [INFO] Session token: a5f8c2d4... (expires in 24 hours)

/help
# Expected: Shows full command list with chat features
```

### Scenario 2: Wrong Password ❌

```bash
# If password is wrong
/login alice wrongpassword
# Expected: ❌ invalid username or password

# Even after failed attempts
/login alice wrongpassword
# Expected: ❌ invalid username or password (no account locked in v2.0)
```

### Scenario 3: Non-Existent User ❌

```bash
# If user doesn't exist
/login nonexistent password
# Expected: ❌ invalid username or password
```

### Scenario 4: Join Room & Chat

```bash
# After successful login
/join general
# Expected: ✓ Welcome to #general

/users
# Expected: [USERS in #general] alice

/msg Hello everyone!
# Expected: Message broadcasted with timestamp

/history
# Expected: Shows last 10 messages with timestamps
```

### Scenario 5: Logout

```bash
/logout
# Expected: ✓ You have been logged out, alice.

# Try protected command after logout
/msg test
# Expected: ❌ You must be logged in to use this command...
```

### Scenario 6: Multi-User Chat

**Terminal 1 (Alice):**
```
/register alice pass123
/login alice pass123
/join general
/msg Hi Bob, welcome!
```

**Terminal 2 (Bob):**
```
/register bob pass456
/login bob pass456
/join general
# Sees: [SYSTEM] alice joined the room
# Sees: [HH:MM:SS] alice: Hi Bob, welcome!

/msg Hi Alice! Nice to meet you.
# Alice sees: [HH:MM:SS] bob: Hi Alice! Nice to meet you.
```

### Scenario 7: Direct Messaging

```
# Alice after logged in
/login alice pass123

# Bob after logged in
/login bob pass456

# Alice sends DM to Bob
/dm bob This is private!
# Bob sees: [HH:MM:SS] [DM from alice]: This is private!

# Bob replies
/dm alice Thanks for reaching out!
# Alice sees: [HH:MM:SS] [DM from bob]: Thanks for reaching out!
```

### Scenario 8: Session Persistence (Within Session)

```bash
/login alice pass123
/join general
/status working
/msg I'm busy right now
# Browser refresh = new connection = new session with same user
```

---

## Edge Cases to Test

### Invalid Usernames

```bash
/register a pass               # Too short (need 2-20)
# Expected: ❌ username must be 2-20 characters

/register alice smith pass     # Contains space
# Expected: ❌ username cannot contain spaces

/register this_is_a_very_long_username_exceeding_limit pass
# Expected: ❌ username must be 2-20 characters

/register alice pass           # Valid, registers if not taken
# Expected: ✓ Account created...

/register alice pass           # Try duplicate
# Expected: ❌ username 'alice' already taken
```

### Invalid Passwords

```bash
/register bob 123              # Too short (need 4+)
# Expected: ❌ password must be at least 4 characters

/register bob password123      # Valid
# Expected: ✓ Account created...

/register charlie <100 chars>  # Way too long
# Expected: ❌ password must be less than 100 characters
```

### Protected Commands Without Auth

```bash
# Without login, try:
/join general
# Expected: ❌ You must be logged in to use this command...

/msg hello
# Expected: ❌ You must be logged in to use this command...

/users
# Expected: ❌ You must be logged in to use this command...

/dm alice hello
# Expected: ❌ You must be logged in to use this command...

# Public commands still work:
/help
# Expected: Shows unauthenticated help

/quit
# Expected: Connection closes
```

### Double Login

```bash
/login alice pass123
# Expected: ✓ Welcome back...

/login alice pass123
# Expected: ❌ You are already logged in. Use /logout first.
```

### Logout Twice

```bash
/login alice pass123
/logout
# Expected: ✓ You have been logged out...

/logout
# Expected: ❌ You are not logged in.
```

---

## Stress Testing

### Test: Multiple Concurrent Users

**Terminal 1:**
```
/register user1 pass1
/login user1 pass1
/join testroom
/msg User 1 here
```

**Terminal 2:**
```
/register user2 pass2
/login user2 pass2
/join testroom
/msg User 2 here
```

**Terminal 3:**
```
/register user3 pass3
/login user3 pass3
/join testroom
/msg User 3 here
```

**Expected:**
- All users see each other's messages
- No session conflicts
- All timestamps accurate
- Room broadcasts work correctly

### Test: Message History

```bash
/login alice pass123
/join myroom

# Send multiple messages
/msg Message 1
/msg Message 2
/msg Message 3
/msg Message 4
/msg Message 5

# View history
/history
# Expected: Shows last 10 (or all 5 if fewer sent)

/history 3
# Expected: Shows last 3 messages

/history 100
# Expected: Shows max available (don't exceed 100 message buffer)
```

---

## Success Criteria Checklist

- [ ] Registration creates accounts with hashed passwords
- [ ] Login validates credentials correctly
- [ ] Logout invalidates session
- [ ] Protected commands require authentication
- [ ] Authenticated users can join rooms
- [ ] Messages display with timestamps
- [ ] Direct messages work between authenticated users
- [ ] User status changes broadcast
- [ ] Room cleanup works (empty rooms deleted)
- [ ] Multiple users can chat simultaneously
- [ ] Message history retrieves correctly
- [ ] Error messages are helpful
- [ ] No race conditions observed
- [ ] Server handles disconnects gracefully
- [ ] Help shows different content (auth vs unauth)

---

## Common Issues & Fixes

### Issue: "You must be logged in" on first command
**Fix:** Register first with `/register username password`, then login with `/login username password`

### Issue: Wrong password error even with correct password
**Fix:** Ensure no typos, passwords are case-sensitive

### Issue: Can't join after login
**Fix:** Make sure you're logged in (check with `/help`), room names are case-sensitive

### Issue: Don't see other users' messages
**Fix:** Both users must be in the same room. Check with `/users`

### Issue: Messages show old timestamps
**Fix:** Server uses system time. Check system clock is correct

---

## Performance Benchmarks (Expected)

- **Registration**: < 100ms (bcrypt hashing)
- **Login**: < 100ms (bcrypt comparison)
- **Protected commands**: < 10ms (session validation)
- **Concurrent users**: Tested with 10+ simultaneous connections
- **Message throughput**: 100+ messages/second

---

## Monitoring Commands (For Testing)

To test server internally:
```bash
# Check what's running on port 8888
netstat -tlnp | grep 8888

# Kill running server if needed
pkill -f TCP-Chat

# Check authentication system is loaded
grep -n "AuthManager" server.go  # Should find several matches
```

---

## Next Steps After Testing

1. ✅ Verify all test scenarios pass
2. ⏳ Add database persistence (SQLite) 
3. ⏳ Deploy to cloud (Heroku/AWS/Azure)
4. ⏳ Set up CI/CD pipeline
5. ⏳ Add logging pipeline
6. ⏳ Create web UI

---

## Quick Copy-Paste Test Sequence

```
# Paste into telnet session:
/register alice password123
/login alice password123
/register bob password456
/join general
/msg Welcome to TCP-Chat v2.0 with auth!
/users
/status working
/history 2
/dm bob Let's test direct messaging
/logout
```

---

**Your authentication system is ready for production testing!** 🎉
