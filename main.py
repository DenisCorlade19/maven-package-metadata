# This is a sample Python script.

# Press Shift+F10 to execute it or replace it with your code.
# Press Double Shift to search everywhere for classes, files, tool windows, actions, and settings.
import subprocess

"""
The following code includes 2 scripts. 
The first script takes a .txt file with groupID:artifactID:version and gets the POM file of the respective package
The second script takes a folder containing just the POM files generated before and uses them to create effective POMS 
which are needed for parsing the metadata.

Comment/uncomment the blocks of code which you need to run.
"""

# Press the green button in the gutter to run the script.
if __name__ == '__main__':


    # First script for generating POM files:

    with open('data.txt') as f:
        readline=f.read().splitlines()
        for row in readline:
            command = f"mvn -Dartifact='{row}' dependency:get"
            subprocess.run([command], shell=True)




    #Second script for generating effective POM files:

    # os.chdir('folder_name')
    # mypath = '.'
    # onlyfiles = [f for f in os.listdir(mypath) if isfile(join(mypath, f))]
    # id = 0
    # for file in onlyfiles:
    #     shutil.copy(file, "pom.xml")
    #     id += 1
    #     effective_command = f"mvn help:effective-pom -Doutput='effective-pom-{id}.xml'"
    #     print(id)
    #     subprocess.run([effective_command], shell=True)

