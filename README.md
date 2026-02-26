![alt text](/assets/icons/icon.png "Logo")

# PasswordSafe
PasswordSafe is an app runnig on Anroid, Linux and Windows. It stores you passords in a SQLite database.
For every entry you give a name (this name is not encrypted and searchable). For every entry you can define fileds (Label and Value) and fill with informations. All this data are stored encrypted.
For encryption AES-256 in conjunction with agon2 is used. The encryption is done with an random Masterkey which is stored encrypted in the database. For opening the masterkey a user Password is used.
You can have multiple categories to store entries in it.
Database can be exported and imported.

PasswordSafe is written in [Go](https://go.dev/) and uses [Fyne](https://fyne.io/) as graphical toolkit.

Author: Reiner Pröls  
Licence: MIT  

## Usage of PasswordSafe
First you have to set a master password. Store it in your brain!  
Then you have to add one or multiple categories. Use the + sign in the toolbar.  
Then you can add entryies to the categories. Use the green + button bottom left to add a new Entry.  
For every entry you can add multiple custom fields. Use the green + button bottom left to add fields.  
For choosing an icon click on the icon.  
On the view Page you can copy the conten of a field in the clipboard. For this use short or long tap. You can define which is used on the settings page.  
You can also yhnage the theme by clicking on the left most icon in the tool bar. Since the original version of the app (15 years ago) was designed for dark theme not all icons will look good in light theme.  

## Screenshots
![alt text](/screenshots/login.jpg "Login screen")
![alt text](/screenshots/categoryview.jpg "Category view")
![alt text](/screenshots/entryview.jpg "Entry view")
![alt text](/screenshots/entryedit.jpg "Edit view")
![alt text](/screenshots/fieldedit.jpg "Field edit")
![alt text](/screenshots/iconselect.jpg "Icon select")
![alt text](/screenshots/searchresults.jpg "Search results")
![alt text](/screenshots/changepassword.jpg "Change password")
![alt text](/screenshots/settings.jpg "Settings view")

### Precompiled binaries
#### Linux (64 Bit)
[Tar file](https://github.com/bytemystery-com/passwordsafe/releases/download/v0.2.6/PasswordSafe.tar.xz)  
[Standalone binary](https://github.com/bytemystery-com/passwordsafe/releases/download/v0.2.6/passwordsafe)  
#### Windows (64 Bit)
[Standalone exe](https://github.com/bytemystery-com/passwordsafe/releases/download/v0.2.6/PasswordSafe.exe)  
#### Mac
Not available - it could be build but requires Mac + SDK.
#### Android 
[APK](https://github.com/bytemystery-com/passwordsafe/releases/download/v0.2.6/PasswordSafe.apk)  

## Q & A
Q: Where is the database stored ?  
>A: Use the Info dialog :-)  
On Linux it will be located at  
~/.config/fyne/com.bytemystery.passwordsafe2/passwordsafe.db
On Windows they are under  
C:\Users\<USERNAME>>\AppData\Roaming\fyne\com.bytemystery.passwordsafe\passwordsafe.db
On Android open the Info dialog at the bottom the path is shown.  
Q: I have forgotten my password ?  
>A: Bad luck ..... nobody can help you. Try to remember !!!


## Links
[Readme](https://bytemystery-com.github.io/passwordsafe/)  
[Repository](https://github.com/bytemystery-com/passwordsafe/)  
[Issues](https://github.com/bytemystery-com/passwordsafe/issues)  
[Discussions](https://github.com/bytemystery-com/passwordsafe/discussions/new)  

© Copyright Reiner Pröls, 2026

