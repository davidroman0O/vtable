
Two part loading system:
- viewport
- bounding checker

We should have some kind of mathematical approach, i have no idea how because i'm not a math girl.

Let's imagine that the viewport can only see 8 items

# when beginning

```
=-=-=-=-=-=-=-=-=- # top bounding loading chunk
================== # top viewport
# - chunk 0
item 0 # cursor
item 1 # top threshold
item 2
item 3 
item 4
item 5
item 6 # bottom threshold
item 7
================== # bottom viewport
# - chunk 1
item 8
item 9
item 10
item 11
item 12
item 13
item 14
item 15
# - chunk 2
=-=-=-=-=-=-=-=-=- # bottom bounding loading chunk
item 16
item 17 
item 18
item 19
item 20
item 21 
item 22
item 23
# - chunk 3 -- no loaded
item 24
item 25
item 26
item 27
item 28
item 29
item 30
item 31
# - chunk 4 -- no loaded
item 32
item 33
item 34
```



# when navigating

```
# - chunk 0 -- no loaded
item 0
item 1
item 2
item 3
item 4
item 5
item 6
item 7
# - chunk 1
item 8
item 9
item 10
=-=-=-=-=-=-=-=-=- # top bounding loading chunk
item 11
item 12
item 13
item 14
================== # top viewport
item 15
# - chunk 2
item 16 # top threshold
item 17 
item 18
item 19 # cursor
item 20
item 21 # bottom threshold
item 22
================== # bottom viewport
item 23
# - chunk 3
item 24
item 25
item 26
=-=-=-=-=-=-=-=-=- # bottom bounding loading chunk
item 27
item 28
item 29
item 30
item 31
# - chunk 4 -- no loaded
item 32
item 33
item 34
```

# when reaching the end


```
# - chunk 0 -- no loaded
item 0
item 1
item 2
item 3
item 4
item 5
item 6
item 7
# - chunk 1
item 8
item 9
item 10
item 11
item 12
item 13
item 14
item 15
# - chunk 2
item 16 
item 17 
item 18
=-=-=-=-=-=-=-=-=- # top bounding loading chunk
item 19 
item 20
item 21 
item 22
item 23
# - chunk 3
item 24
item 25
item 26
================== # top viewport
item 27
item 28 # top threshold
item 29
item 30
item 31
# - chunk 4 -- no loaded
item 32
item 33 # bottom threshold
item 34 # cursor
================== # bottom viewport
=-=-=-=-=-=-=-=-=- # bottom bounding loading chunk
```

