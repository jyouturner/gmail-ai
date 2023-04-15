# Polling Messages from Source and Process

This is the process that periodically poll messages from the source (ie, Gmail) and process the messages. 

It includes a concret polling process and a bunch of interfaces ans data structure:

* Message is generatic struct with ID, Subject, Body etc.

* MessageService interface defines the methods to provide messages, and ids.

* MessageHandlerFunc for message handling logic (for example to detect rejections and add label) to implement.

