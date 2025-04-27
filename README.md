# Messaging CLI: Go and TView

A CLI-based real time messaging application allowing users to sign up and chat with other users ona local network in the terminal.

Features a notification bar and a game section (currently only Snake is available).

I undertook this project to learn Golang as well as to get familiar with using websockets, SQLite, and the Tview CLI builder library in Go.

At the moment, the application functions on a Debian 11 machine and has not been tested for compatibility with other operating systems. As well, I have not implemented a mechanism to 
automatically detect the subnet that the server is located on - this will need to be configured 
manually at this point. 


## Goals
  - Create a functional Real-time messaging app in the CLI using  websockets and client-server architecture.
  - Effectively implement Goroutines and channels to manage concurrent aspects of the application. 
  - Implement CRUD functionality with some complex state management across different components of the CLI app -- using SQLite.

## Demo Video

-[Create first schema](https://vimeo.com/1072854509)

-[Capturing content from Companies House](https://vimeo.com/1072854487)

-[Using multiple schemas](https://vimeo.com/1072854500)



## Roadmap

- Initial Research:

  - Brainstorm features, use-cases, challenges, project structure, software design choices, testing strategy etc.
  - Recorded technologies to use. These were:
    - **TypeScript**
    - **Vite**
    - **Webpack**
    - **Jest**
    - **Chrome API**
    - **React**
    - **React Router**
  - Mapped out a data storage model which complements IndexedDB.
  - Data structures:
    - **User Content Model**: object containing all user content (projects, schemas, captures, user data).
    - **Project**: object representing a single project. Projects contain captures, which are schema models with additional data relating to the capture context (url, data etc). Captures must belong to a project.
    - **Schema Model**: object representing a single schema. Schemas contain entries mapping match expressions to values. The content script parses these schemas to locate and extract data from the DOM.

- UI Design

  - Little bit of research into what good Chrome Extension UI design looks like.
  - Work out action flow -- what will user see, what actions should they perform, what options should be exposed and where do I present them?

- UserContentModel controller class implementation.
- UI design and build.
- Service Worker/background design and implementation.
- Scraping functionality implementation and side panel for visualising extracted data.
- Miscellaneous features and bug fixes.
- Looking forward.

## Initial Research

### Features

- Create schemas by highlighting and clicking elements on a webpage. The schema can be reused on other webpages with similar urls to then capture data corresponding to where the user clicked when creating the schema. Schemas should hold the following data:

  - Url pattern to match with webpages.
  - Match expressions (e.g. a CSS selector), match expression types (i.e. is an id, or a generic CSS selector, or manual input), and matched values.

- Automatic url matching mechanism, highlighting to user when there is webpage that can be scraped.
- Shortcut keys for capture pages, editing templates, and creating schemas.
- Update, delete, create, and review schemas, and captures.
- Create projects and hold captures under separate projects.
- Persist data between sessions.
- Manually populate data fields in a capture by clicking elements on a page.
- Export the data in a JSON format for ease-of reuse
- User can provide CSS selectors (including ids) to match for elements on the page if desired.
- Side panel for more space to work with data collected from captures.

### Use-cases

- Corporate Intelligence context,for example. Manually collecting data from frequently visited webpages can be time consuming and error prone. Automatic scraping tools can suffer from a lack of flexibility or an inability to deal with dynamically generated content (i.e. generated when interacted with by user), or not being able to bypass Captcha security mechanisms.

- The extension should allow a researcher to easily capture a web page by using a pre-defined schema which matches with various elements on the page, bundle captures into projects, and export the captures or projects.

### Project Structure

- I chose to split my project into each logical Chrome Extension component: popup, side panel, background (service worker), content script, and shared content. Each directory has its own webpack or Vite build config.

- There may have been a better way to bundle a project like this together, but I wanted to separate the code into logically distinct parts.

- There is a types directory for sharing TypeScript types across all parts of the project. This helped maintain consistency with working with data structures.

### Software Design Patterns

- I wanted to implement something similar to a MVC-architecture. The Model in this case being IndexedDB.

- I then designed a class called UserContentModel, accessible by the GUI parts of the Chrome Extension. This is analogous to the controller in MVC. The class is a Singleton which holds virtual copy of all the user content in the database, and also communicates with the service worker to perform operations on the database. The class reflects any changes in the local virtual model.

- The motivation behind creating a class like this was the following:

      * Creating virtual copy of user content held throughout the duration of the GUI lifetime means less trips are necessary to the IndexedDB store to search for user data. This is important because IndexedDB operations in my application is being handled by the background script. Therefore messages need to be sent across the extension/background script boundary to execute any database operations. In theory, I am saving some time and resources by simply querying a model in memory rather than constantly fetching and creating new objects constantly.

      * I also attached event listeners to carry out operations. I exposed the methods related to triggering the event listeners rather than exposing the user content model itself. I'm not sure if this is optimal, but I wanted to experiment with OOP principles, especially encapsulation, and also the event emitter software design pattern. I also implemented a Singleton design pattern -- only the event listener method is exposed, leaving a single UserContentModel to persist.

### Testing

- I began writing the UserContentModel class using a test driven approach. I created around 40 unit tests to check that the class reacts correctly to events emitted on it, and correctly updates the virtual UserContentModel when IndexedDB operations complete successfully. I also mocked some Chrome API methods to simulate the flow of operations after triggering an event listener, such as sendMessage and onMessage methods.

- I stopped writing tests because of the added complication of requiring mocks, which slowed down development. That being said, I found a test driven approach to be really satisfying as it helped me visualise the progress I was making, and made me think explicitly about the shape of the application functionality.

- I used the jest framework. I used mocks as recommended by Chrome docs - https://developer.chrome.com/docs/extensions/how-to/test/unit-testing

- I declared a setup file in a jest config file, we can define code to create the mock apis

### React Router

- Using React Router for client side routing. Easy to implement although I think it's not necessary.

## UI Design

### UI Ideas Research

- Examples of Chrome Extension designs I drew inspiration from:

- Window Resizer
- Mota: https://dribbble.com/shots/23226306-Mota-Browser-extension-UX-UI-design-for-the-remote-work#
- Comparify: https://dribbble.com/shots/24520989-Comparify-Price-Comparison-Extension

- WebasLoad:https://dribbble.com/shots/24580883-WEBASLOAD-Chrome-Extension
- https://dribbble.com/shots/22955117-Tetua-AI-Chrome-Extensions
- https://dribbble.com/shots/22004480-Time-Tracker-Widget-macOS-timebite

- https://dribbble.com/shots/10689249-Omni-Card-Creation

### User Interaction Flow

- I created slightly different designs for the popup and the side panel. The popup contains most of the functionality for reviewing content, exporting data, and managing projects. The side panel is reserved for managing content generated by the schemas.
- The action:
  - Welcome screen. Provides short introduction and link to create schema and getting started.
  - Home screen
    - Current project (view project and set new project).
    - Schema matches (list of schemas usable on the current webpage).
    - Options to view projects and schemas.
    - Primary action to scrape page.
  - Content views (projects, schemas, captures) with operations, edit, save, delete.
  - Project list --> project view. Project view allows you to check captures.
- Side Panel
  - Three views for editing captures, creating schemas, and editing schemas.

## Background

- I created a extension messaging type BackgroundMessage, which carries information about the desired operation and associated data. E.g. database, openSidePanel, getCurrentTab.

### Functionality

- Create context menu options.
- Listen out for shortcut key events and trigger appropriate response.
- Create database and manage data persistence via IndexedDB API.
- Detect when current tab and url changes, and find schemas matching current url. Then trigger action badge if any match.

## Side Panel

- The side panel is used when a user wants to edit or create a schema, or edit a capture from a page.
- The side panel does not use the UserContentModel as the popup does. Instead, it manages a "Schema Model", or a representation of the Schema or Capture the user is currently working with.
- Like the UserContentModel, the Schema Model is a local object to the side panel and does not affect the database, unless the user saves any changes made.
- I implemented a custom hook - useModel - which manages the local representation of the Schema Model, as well as data relating to the state of the row and cell selection in the side panel.
- The useModel hook returns three items: the schema model (formModel), a reducer helper object, and the id of the currently focused cell.
- Every time the schema model is modified via the reducer helper object, the changes are propagated via the Context API. All components have access to the schema model, reducer object, and focused cell.
- The reducer object exposes several methods, which when called, updates the schema model with any content changes.
- When the user saves changes, this triggers a database event in the background script, saving the new content to the database.

### Content Script and DOM Scraping

- When the user triggers a capture or edit schema event, the background script opens the side panel and the sends the matching schemas to it. The side panel then establishes a connection with the content script, sends the schema over the connection, and awaits the results of the scrape operation from the content script. The matched data then populates a table in the side panel which the user can edit.

- The content script is injected into all webpages, and awaits a connection from the chrome extension before it performs any tasks.

- Once a connection is established with the content script, the content script listens for two types of requests: a fetchMany or fetchOne request.

  - **FetchMany**: Content script receives an entire schema, containing match expressions (CSS selectors or ids) and retrieves matched data.
  - **FetchOne**: Content script attempts to find a text node selected by the match expression provided by the side panel. This can be an id or a CSS selector.

- The content script also listens for user click events. When a user clicks over a text node, the content script obtains the text node underneath the cursor, builds a selector, and sends the data to the side panel. The side panel then assigns the fetched data to the currently highlighted cell in the table.

### Multiple Schema Matches

- If a user navigates to a page where there are multiple schemas with matching urls, then triggering a capture or edit schema operation will take one of the matched schemas and forward that to the side panel.

- The side panel then picks one of the schemas, triggers a content script fetchMany event, and displays the data in a table. The user can switch between any schema using a dropdown. Each time the user switches between schemas, the side panel triggers a new fetchMany event.

### Triggering the side panel

- The side panel can be opened via multiple methods. The user can open it manually when navigating the popup, and chooses to edit a schema.
- The user can also take advantage of context menu buttons and shortcut key (CTRL + SHIFT + 1 - 3).

## Miscellaneous features

- **Export Content**: Captures and projects can be exported to a JSON file via the popup.
- **Text Highlighting**: While the side panel is open, the hovered text is highlighted.
- **Custom CSS selector scheme**: The CSS selector built by the content script cycles through the parents of the selected text node, retrieving id and class names, and constructs a string with that data. As child nodes have multiple sibling nodes, the content script encodes the index of the child node to look for in each layer of the resulting CSS selector. E.g. div#dummy child:3; h1 child: 0; indicates that for the div#dummy element, the content script must select the 3rd child node.

## Next Steps

- **Support for selecting tables**: tables present a challenge with the current text node matching algorithm. I will look to add support for selecting tables, which often have variable sizes.

- **Improve matching algorithm**: The current matching algorithm relies on the positions of child nodes to locate a desired text node. This is an issue for pages where the structure is not totally fixed, e.g. for company information there might be lines added or omitted from the page depending on its status, the type of company it is, and so forth. When a user visits another company page, the schema they created might not match all the elements in the new page because of the difference of the relative text node placements between first and the new page. Currently, the user can remedy missing entries by clicking the missing content, which will populate empty records with newly fetched records, but this might be cumbersome over time.

- **Testing**: Update test suite to cover the whole application, and introduce end-to-end testing.

- **Develop extension for Microsoft Edge**

- **Synchronise data across browsers**