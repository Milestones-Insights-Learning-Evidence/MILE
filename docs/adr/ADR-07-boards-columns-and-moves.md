# ADR-07: Boards, columns, moves, blocking and task closure

- Status: Accepted, 15 November 2026. Revised the same day after the schema audit. Section 13 lists what still needs confirming.
- Decisions: DEC48, DEC49, DEC50, DEC53, DEC66, DEC69, DEC78, DEC79, DEC80, DEC81 (decisions workbook v8)
- Plan: implementation plan v4.0, sections 3.2, 3.3 and 5.1; needed before the Phase 2 schema

## Context

The backward-move prompt, the confirmation queue and every later evidence calculation depend on what counts as a move. The plan asks this ADR to fix boards, column order, terminal columns, blocking, same-column drops and column archival before the schema is written. Four further rules were accepted on 15 November (DEC78 to DEC81). An audit of the first schema then found fourteen ways around the rules; the fixes are part of this revision.

## Decision

### 1. Boards

- A board belongs to exactly one cohort, and a cohort has exactly one board (DEC48).
- A board follows one workflow definition of the cohort's department. At any moment it uses one version of that definition, and it only ever moves to a newer version. Each change of version is recorded as an event.
- Mentors filter the board by intern. What each role may see follows DEC68.

### 2. Workflow versions are immutable

- A workflow definition has numbered versions. A version is a fixed, ordered list of columns.
- Any edit (add, remove, rename or reorder a column, change a terminal flag or a dwell limit) creates a new version. A version and its columns are written in one transaction. After that nothing can be added, changed or removed, so a recorded event keeps its meaning.
- Each column carries a `column_key` that stays the same across versions, so a renamed column can be recognised as the same column.
- A version needs at least one terminal column and at least one non-terminal column.

### 3. Moving a board to a new version (DEC78)

- The administrator supplies a map from every column of the old version that holds cards to a column of the new version.
- One transaction switches the board to the new version, records a board-version event, moves each open card to its mapped column, raises each task's version by one, and records a `remap` transition for each card. The database refuses to commit a switch that leaves an open card behind.
- A `remap` transition is never backward, needs no reason, cannot be undone, and is not counted as a move. Evidence calculations treat it as a continuation of the time in the previous column (worked examples in Phase 6).
- Whether a later move is backward is judged by column order in the version in force at the time of that move.
- An abandoned task is not remapped. It keeps its last column, in the version it was closed under.

### 4. What counts as a backward move (DEC49)

A `move` is backward when, in the version in force:

- the target column comes earlier than the source column, or
- the source column is terminal and the target is a different column, including another terminal column.

A forward move of any distance needs no reason. A backward move is held until the mover picks a reason; the API returns `428` and nothing is committed.

### 5. Removing a column (DEC79)

- A new version can leave out a column only if no card sits in it, or if the map in section 3 sends its cards to another column.
- The column stays in every older version, so history still reads correctly. Nothing is deleted.

### 6. Same-column drops and reordering inside a column (DEC80)

- Dropping a card back into its own column does nothing: no event, no version change.
- Reordering cards inside a column records no event and does not raise the task version, so it can never cause a `409` against a real move.
- A card's position is a separate rank value on the task. The last write wins.

### 7. Blocking (DEC49, DEC69)

- Blocking is a flag on the card, not a column. Setting it and clearing it are recorded as events, and the pairs form block intervals.
- Setting a block needs a free-text note, because a block can be caused by anything unrelated to the intern's skills. Clearing needs no note.
- A block set by an intern is excluded from cycle time only after a mentor approves it, under the DEC53 rules: any mentor of that intern, a reminder after 2 working days, a department administrator after 5. A rejected block stays in the history, and its time counts as normal. The task's own intern can never approve.
- A block set by a mentor or administrator is approved in the same transaction.
- Setting or clearing a block raises the task version. A blocked card can still be moved.
- A block still open when a task is abandoned ends at the closure.
- Plan default D14 counts each time the flag is set.

### 8. Scope changes

- A mentor or administrator can mark a scope change on a task; the task's own intern cannot. It is recorded as an event, takes no note (DEC66) and raises the task version.

### 9. Ending a task without finishing it (DEC81)

- A mentor or administrator can close a task as abandoned, with a required note; the task's own intern cannot. It is recorded as an event and the task is never deleted. The closing time is the database's.
- A closed task cannot be moved, blocked, reassigned, retagged or edited.
- A task in a terminal column cannot be closed as abandoned, since it is already finished.
- An abandoned task is left out of the completed-task count.

### 10. Tags and text edits

- Skill tags decide which skill a counted reason is attributed to, so tag changes are recorded as add and remove events and the history is kept. A task always has at least one tag (D7).
- Every edit of a task's title, description, external reference or time allowance is recorded: who, when and which fields. The old text is not kept, so no extra copy of typed text exists (DEC66).
- Both raise the task version.

### 11. Creation, assignment, reasons and undo

- Creating a task records an `enter` transition into the chosen column, so first-column time has an anchor. A task is created in a non-terminal column, with its first tag and its assignment.
- Every task has one intern assignee, a membership in the board's cohort. Each change of assignee is recorded, and every event stores the assignee at that moment as its subject (DEC50, DEC61).
- The reason list is not the client's choice. If the mover is the task's own intern, the intern list applies; anyone else who may move the card is a mentor or administrator, and the mentor list applies (DEC10, DEC11).
- An intern can never settle their own reason. An intern is never a mentor of their own membership.
- An undo is a transition of kind `undo` back to the source column, linked to the move it reverses. Only the mover may undo, inside the undo window (1 minute by default, DEC39), and only while the task is unchanged since the move. Otherwise the API returns `409`. An undo needs no reason and cancels any pending confirmation.

### 12. Concurrency and the write order

- Every change to a task raises its version by exactly one, except reordering inside a column. A request carries the version it expects, and a stale request gets `409` with the current task.
- A change is written in one transaction: first the task row, with `WHERE version = <expected>`, then the event row or rows. The event triggers lock the task row, so two writers on one task take turns.
- Every version of a task has at least one event row, and each thing that changed has the event that records it. The database refuses to commit otherwise.
- A task write takes a shared lock on its board, so a version switch and a card write cannot cross. When the two meet, PostgreSQL may report a deadlock or a serialization failure; the application retries the request.
- Event times come from the database clock at the moment of the write, and never run backwards within one task.

### 13. To confirm

These rules were chosen to fill gaps and have not been confirmed one by one (O60):

1. A closed task can never change.
2. A task in a terminal column cannot be abandoned.
3. An abandoned task cannot be reopened in the first release.
4. A blocked card can still be moved.
5. Setting or clearing a block raises the task version.
6. A move from one terminal column to another counts as backward.
7. Every task has its intern assignee from creation; there is no unassigned backlog.

One question is open (O63): the schema lets the mover undo any move inside the window, forward or backward. DEC14 and DEC39 speak only of backward moves. Undoing a forward move sends the card back without a reason, which is the cheapest honest fix for a slip, and both events are kept.

## Consequences

The Phase 2 schema enforces these rules in the database, not only in the application. `db/SCHEMA.md` lists each rule with the mechanism that enforces it, and `db/test/schema_test.sh` holds a test for each.

The database knows who a task's own intern is, but it does not know roles until Phase 3. Rules of the form "a mentor of that intern" or "a department administrator after 5 working days" stay in the application and the permissions matrix.

Open for later: reopening an abandoned task, and holidays in the working calendar (only the working week is configurable today).
