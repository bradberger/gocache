echo "mode: atomic" > coverage.txt
for Dir in $(find ./* -maxdepth 10 -type d );
do
        if ls $Dir/*.go &> /dev/null;
        then
            go test -coverprofile=profile.out -covermode=atomic $Dir
            if [ -f profile.out ]
            then
                cat profile.out | grep -v "mode: atomic" >> coverage.txt
            fi
fi
done

rm -Rf *.out
