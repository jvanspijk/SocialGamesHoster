- During playing the Echo Location game: the lookout was unable to use the lookout chat. This same issue might also occur for the other roles.

---

- approvals require leaving the lobby, this is not a bug but there should be a direct workflow for the game owner to accept/deny approvals easily from the live game screen.
- Logging out, creating a new account and then logging in with the old name creates a new profile instead of logging in with the old profile (unsure if this is true, might have to do with case sensitivity. Given that the displayed names are capitalized, this should not be case sensitive.)
- Players can't join a game after it's already started

---

- There's no distinction between application errors and validation errors when displaying errors. This is confusing for users as they are reporting errors when for something simple as a wrong password input or invalid profile name.
	- Validation errors should also not create a trace log. 
	- DO NOT solve this by appending Validation error: <error message>. This would look highly unprofessional and the user does not need to know the type of error.
