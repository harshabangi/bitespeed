# bitespeed

This is a Golang backend hosted at https://bitespeed-01lh.onrender.com/swagger/index.html#. The endpoint is also accessible at https://bitespeed-01lh.onrender.com/identify.

## Main Idea
The main top-level idea behind the solution for the problem statement is as follows:

1. Consider each email and phone number as a node in a connected component.
2. If both the email and phone number are not present in any connected component, a new connected component with new nodes is created.
3. If an incoming email or phone number is already present in a connected component, and the other (email or phone number) is new, the new email or phone number is added to the existing connected component.
4. If both the incoming email and phone number are present in two different connected components, these connected components are merged into one.
