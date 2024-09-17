# Project SAT FORM

## Description

This project is a simple web application in golang that allows users to do the following:
- Submit SAT Results and assign a rank to the record based on the SAT score
- View them all in a table
- Get a rank based on the name provided
- Update the SAT score of a record
- Delete a record
- A view all data button to view all data in the database in JSON format

## Database Design

### SAT Score Table
- id (unique)
- name (unique)
- address
- city
- country
- pincode
- sat_score
- passed (boolean, true if sat_score is greater than 30% else false)
- created_at
- updated_at
- rank

## Technologies

- Backend: Golang
- Database: SQLite
- Frontend: Templ, TailwindCSS, HTMX, AlpineJS

## Plan

- [X] Setup http server to handle requests, and define the necessary handlers
- [X] Setup the database and create the necessary tables
- [X] Setup the empty frontend
- [ ] Implement that handlers one by one
    - [X] Submit SAT Score
    - [ ] Delete SAT Score
    - [ ] View all data in JSON format
    - [ ] Search for a SAT Score by name
    - [ ] Update SAT Score
- [ ] Add tailwind