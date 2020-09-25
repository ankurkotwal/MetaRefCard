# HotasRefCard

# Plan
1. ~~Read FS2020 xml files~~
2. ~~Read the EDRefCard inputs~~
3. ~~Build a model of game inputs and controller mappings~~
4. ~~Generate images~~
5. ~~Dynamic font size~~
6. ~~Regenerate hotas_images, new X55 locations, vkb-kosmosima-scg-left 3879x2182, x-45 5120x2880~~
7. ~~Convert to webapp~~
8. ~~Sliders~~
9.  ~~Add game banner~~
10. Colours
11. Make images clickable to open a new tab
12. Add Google Analytics
13. Build container image
14. Publish on Cloud Run
15. Keyboard & mouse
16. Extend to Elite Dangerous


# Setup

## Python
### Generate Device Model
#### Dependencies
Install modules
```pip3 install pyyaml```
#### Running the script
Read `3rdparty/edrefcard/bindingsData.py` to generate a custom configuration.
Command:
```generateControllerInputs.py```

# Generate Hotas Images
#### Dependencies
* Inkscape
* Imagemagick
```pip3 install ```
#### Running the script
Generate jpgs of the Hotas images found in `assets/hotas_images` into `refcard/resources/hotas_images`
Command:
```generateHotasImages.py```
