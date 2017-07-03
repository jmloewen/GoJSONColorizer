package main
import(
    "os"
    "bufio"
	"io/ioutil"
    "strconv"
    "fmt"
)

/*
    Written by Jonathan M. Loewen STN 301281915 for CMPT-383 Comparative Programming Languages, Summer 2017 SFU
    Principal coding: Week of May 22nd, I believe.  Maybe a bit before that.
    Completed: June 5th, 2017.
    
    June 3rd Comments:
    I did not use a tokenizer initially, as it was not a strict requirement for this assignment.
    What i instead did was to set up a series of flags indicating what I was parsing, and then color the parsed values appropriately.
    I have a large number of functions, some of which are somewhat superfluous, but it seemed inappropriate to have functions for only a few cases.
    
    This could be easily condensed into fewer functions (and probably more code), but I felt that it was more clear with more functions.  As a
    result, this is not clearly delineated in 'Is Tokenizer equivalent' and 'Is I/O', but it should be very clear what's happening here.
    
    Update June 5th:
    I decided to add a tokenizer, but I don't color based on this tokenizer, because that's slow and useless.  I did tokenize according to the assignment guidelines though,
    and it passes my test cases of 'throw a whole bunch of garbage characters into the tokenizer and see what happens'.
    
    This operates on the assumption of valid JSON, and will cover a couple of invalid cases, but clearly not all of them.
    
    -Jonathan
*/
//-----------------------------------------------------------------------------------
///Area for global variables - I probably didn't need as many as I used, but oh well.

type Token interface{}
type TokenValue uint8
const(
    uncolored  TokenValue = 0
    curlyBrace TokenValue = 1
    squareBrace TokenValue = 2
    colon TokenValue = 3
    comma TokenValue = 4
    str TokenValue = 5
    boolean TokenValue = 6
    escapeStr TokenValue = 7
    chunk TokenValue = 8
    unknown TokenValue = 9
)

//the number of tabs that should be printed out after a newline.
var numOfTabs = 0

//flags indicating what type of data segment we are in.
var isInString = false
var isInChunk = false
var isInBool = false
var isInEscapeString = false
//a counter for how many characters have occured in an escape string.
var escapeStringCounter = 0
var sliceOfTokens []Token

//-----------------------------------------------------------------------------------
///This area is for formatting and shaping of the JSON.  These functions include adding tabs, opening and closing tags, and colorization.

//Add the appropriate amount of tabs to the current line.  This does not increment our number of tabs.
func addTabs()string{
    tabString := ""
    for i := 0; i < numOfTabs; i++{
        tabString += "&nbsp;&nbsp;&nbsp;&nbsp;"
    }
    return tabString
}

//turn on our colorization for the segment.
func startSpanning(colorVal int)string{
    colorString := "<span style="
    switch colorVal{
        //curly braces
        case 0:
        colorString+="color:magenta"
        //square braces
        case 1:
        colorString+="color:green"
        //Colons
        case 2:
        colorString+="color:orange"
        //Commas
        case 3:
        colorString+="color:red"
        //Strings
        case 4:
        colorString+="color:blue"
        //Booleans
        case 5:
        colorString+="color:olive"
        //Numbers / 'Chunks'.
        case 6:
        colorString+="color:gray"
        //Escaped Strings.
        case 7:
        colorString+="color:cyan"
        default:
        //this is our error value.
        colorString+="color:maroon"
    }
    return colorString + ">"
}

//turn off our coloization for he segment.
func stopSpanning()string{
    return "</span>"
}

//-----------------------------------------------------------------------------------
///This area is for 'Toggle' functions to move us in and out of 'String', 'Boolean', 'Chunk', and 'Escaped String' modes.
///It is nicer to have functions like this than to have to toggle our booleans every single time we call them.  This way, the process
///becomes standardized and reduces confusion.

//a string is anything surrounded by non-escaped " marks.
func toggleStringStatus()string{
    if isInString{
        isInString = false
        return stopSpanning() + "&quot;"
    }else{
        isInString = true
        return "&quot;" + startSpanning(4)
    }
}

//an escape stringis any string that follows a backslash.
func toggleEscapeStringStatus()string{
    if isInEscapeString{
        isInEscapeString = false
        escapeStringCounter = 0
        return stopSpanning()
    }else{
        isInEscapeString = true
        escapeStringCounter = 1
        return startSpanning(7)
    }
}

//A Boolean is a true/false/null value that is not in quotes.
func toggleBoolStatus()string{
    if isInBool{
        isInBool = false
        return stopSpanning()
    }else{
        isInBool = true
        return startSpanning(5)
    }
}

//A Chunk is any value that is outside of strings, but a valid value (AKA: Numbers, not in quotes)
func toggleChunkStatus()string{
    if isInChunk{
        isInChunk = false
        return stopSpanning()
    }else{
        isInChunk = true
        return startSpanning(6)
    }
}

//-----------------------------------------------------------------------------------
///This area is for dealing with characters when we're in a specific 'mode', after having fallen into the Default.
///it includes functionality for handling escapes, booleans, numbers, and strings.

//We're in an escape string - deal with this outcome.
func handleEscapeStrings(s string)string{
    _, err := strconv.Atoi(string(s[0]));
    //if its a number, increase our counter.
    if err == nil{
        if escapeStringCounter == 6{
            return s + toggleEscapeStringStatus()
        }else if escapeStringCounter == 2{
            return s + toggleEscapeStringStatus()
        }else{
            return s
        }
    //if its a u, then we have to kick off our u#### escape string.  Otherwise, we need to close our escape string situation.
    }else if string(s[0]) == "u" && escapeStringCounter == 2{
        return s
    }
    //else, we're not in a u#### situation, we're in a \n or \r situation. add the character and toggle escape off.
    return s + toggleEscapeStringStatus()
}

//Start up a numeric / 'chunk' span area.
func handleChunks(s string)string{
    return toggleChunkStatus() + s 
}

//Start up a boolean span area.
func handleBooleans(s string)string{
    return toggleBoolStatus() + s
}

//handles the Default area of our main switch case.
func handleDefaultCase(s string, i int)string{
    if (isInString && !isInEscapeString) || isInChunk || isInBool{
        if isInString{
            sliceOfTokens[i] = str
        }else if isInChunk{
            sliceOfTokens[i] = chunk
        }else{
            sliceOfTokens[i] = boolean
        }
        return s
    }
        
    //if we're not just entering an escape string, decide what data we're parsing with our escape.
    if escapeStringCounter > 0{
        sliceOfTokens[i] = escapeStr
        return handleEscapeStrings(s)
    }
    
    //can we enter a chunk?  If this does not error, it means we have numbers.
    _, err := strconv.Atoi(string(s[0]));
    
    //if err is nil (or we find a '-'), this is a number which means we're now in a chunk.
    if err == nil || string(s[0]) == "-"{
        sliceOfTokens[i] = chunk
        return handleChunks(s)
    }
    
    //Finally, we're not in a string and we're finding letters as values - are they 't', 'f', or 'n'? then we've encountered a boolean.
    if string(s[0]) == "t" || string(s[0]) == "f" || string(s[0]) == "n"{
        sliceOfTokens[i] = boolean
        return handleBooleans(s)
    }
    
    //if we're here, we could not enter an escape, a chunk, or a bool and are not in a string. Something has gone wrong - the json is likely invalid.  Retain the null.
    fmt.Println("Encountered uncovered situation - Cannot properly parse, sorry!")
    return startSpanning(10) + s + stopSpanning()
}

//-----------------------------------------------------------------------------------
///This is our main formatting area, which decides what to do with what types of characters, if we're not in a specific 'mode'.
///Extra code (like in , }, [, and ]) is for formatting purposes, and most easily done here rather than up above.
func formatCharacter(s string, i int)string{
    //if we're in an escape string, update the counter so that later we will know when to leave it.
    if isInEscapeString {
        escapeStringCounter++
        sliceOfTokens[i] = escapeStr
    }
    //switch case for specific cases
    switch s{
        case "{":
            if !isInString{
                s= "<br/>" + addTabs() + startSpanning(0) + s + stopSpanning() + "<br/>"
                numOfTabs++
                s += addTabs()
                sliceOfTokens[i] = curlyBrace
            }else if isInEscapeString{
                sliceOfTokens[i] = escapeStr
                return s + toggleEscapeStringStatus()
            }else{
                sliceOfTokens[i] = str
                return s
            }
        case "}":
            if !isInString{
                if isInBool{
                    toggleBoolStatus()
                }else if isInChunk{
                    toggleChunkStatus()
                }
                numOfTabs--
                s= "<br/>" + addTabs() + startSpanning(0) + s + stopSpanning()
                sliceOfTokens[i] = curlyBrace
            }else if isInEscapeString{
                sliceOfTokens[i] = escapeStr
                return s + toggleEscapeStringStatus()
            }else{
                sliceOfTokens[i] = str
                return s
            }
        case "[":
            if !isInString{
                s= "<br/>" + addTabs() + startSpanning(1) + s + stopSpanning() + "<br/>"
                numOfTabs++
                s += addTabs()
                sliceOfTokens[i] = squareBrace
            }else if isInEscapeString{
                sliceOfTokens[i] = escapeStr
                return s + toggleEscapeStringStatus()
            }else{
                sliceOfTokens[i] = str
                return s
            }
        case "]":
            if !isInString{
                if isInBool{
                    toggleBoolStatus()
                }else if isInChunk{
                    toggleChunkStatus()
                }
                numOfTabs--
                s= "<br/>" + addTabs() + startSpanning(1) + s + stopSpanning()
                sliceOfTokens[i] = squareBrace
            }else if isInEscapeString{
                sliceOfTokens[i] = escapeStr
                return s + toggleEscapeStringStatus()
            }else{
                sliceOfTokens[i] = str
                return s
            }
            
        case ":":
            if !isInString{
                s= startSpanning(2) + s + stopSpanning()
                sliceOfTokens[i] = colon
            }else if isInEscapeString{
                sliceOfTokens[i] = escapeStr
                s = s + toggleEscapeStringStatus()
            }else{
                sliceOfTokens[i] = str
            }
            return s
        case ",":
            var closeOldTags = ""
            if !isInString{
                if isInBool{
                    closeOldTags = toggleBoolStatus()
                }else if isInChunk{
                    closeOldTags = toggleChunkStatus()
                }
                s= closeOldTags + startSpanning(3) + s + stopSpanning() + "<br/>" + addTabs()
                sliceOfTokens[i] = comma
            }else if isInEscapeString{
                sliceOfTokens[i] = escapeStr
                return s + toggleEscapeStringStatus()
            }else{
                sliceOfTokens[i] = str
                return s
            }
        case "<":
            s="&lt;"
            if isInEscapeString{
                sliceOfTokens[i] = escapeStr
            }else if isInString{
                sliceOfTokens[i] = str
            }else{
                sliceOfTokens[i] = uncolored
            }
        case ">":
            s="&gt;"
            if isInEscapeString{
                sliceOfTokens[i] = escapeStr
            }else if isInString{
                sliceOfTokens[i] = str
            }else{
                sliceOfTokens[i] = uncolored
            }
        case "&":
            s="&amp;"
            if isInEscapeString{
                sliceOfTokens[i] = escapeStr
            }else if isInString{
                sliceOfTokens[i] = str
            }else{
                sliceOfTokens[i] = uncolored
            }
            
        case "\n", "\r":
            //nothing to tokenize, we're returning an empty string, but we need to assign it a value.
            s=""
            if isInEscapeString{
                sliceOfTokens[i] = escapeStr
            }else if isInString{
                sliceOfTokens[i] = str
            }else{
                sliceOfTokens[i] = uncolored
            }
        case "\\":
            if !isInEscapeString{
                s = toggleEscapeStringStatus() + s    
            }else{
                s += toggleEscapeStringStatus()
            }
            sliceOfTokens[i] = escapeStr
            return s
        case "\"":
            if !isInEscapeString{
                s = toggleStringStatus()
                sliceOfTokens[i] = uncolored
            }else{
                s = "&quot;"
                sliceOfTokens[i] = escapeStr
            }
        case "'":
            s= "&apos;"
            if isInEscapeString{
                sliceOfTokens[i] = escapeStr
            }else if isInString{
                sliceOfTokens[i] = str
            }else{
                sliceOfTokens[i] = uncolored
            }
        case " ":
            //sliceOfTokens[index] = 'space'
            s= "&nbsp;"
            if isInEscapeString{
                sliceOfTokens[i] = escapeStr
            }else if isInString{
                sliceOfTokens[i] = str
            }else{
                sliceOfTokens[i] = uncolored
            }
        case "\t":
            s= "&nbsp;"
            if isInEscapeString{
                sliceOfTokens[i] = escapeStr
            }else if isInString{
                sliceOfTokens[i] = str
            }else{
                sliceOfTokens[i] = uncolored
            }
        default:
        //we've hit the default - go to the function that deals with this.
        return handleDefaultCase(s, i)
    }
    //if we've hit 2 characters in our escape string, turn it off unless that character is a u (in which case we have 4 numbers ahead of us).
    if escapeStringCounter == 2{
        //the only acceptable 2nd character for an escape if we want to continue is 'u'.
        if s != "u"{
            sliceOfTokens[i] = escapeStr
            return s + toggleEscapeStringStatus()
        }
    }
    
    if sliceOfTokens[i] == nil{
        fmt.Println("Encountered uncovered situation - Cannot properly parse, sorry!")
        return startSpanning(10) + s + stopSpanning()    
        if isInEscapeString{
            sliceOfTokens[i] = escapeStr
        }else if isInString{
            sliceOfTokens[i] = str
        }else if isInChunk{
            sliceOfTokens[i] = chunk
        }else if isInBool{
            sliceOfTokens[i] = boolean
        }

    }
    
    return s
}

//-----------------------------------------------------------------------------------
///This mini-area contains two functions for setting up the page.  One sets up the initial state for the page, the other calls the above functions.

//our main parser for taking json and passing it to our stylize functions.
func stylizeHtmlPage(dat []byte)string{
    
    sliceOfTokens = make([]Token, len(dat))
    stylizedData := ""
    //we have all of our data - it's time to parse it.
    for i:=0; i < len(string(dat)); i++{
        stylizedData += formatCharacter(string(dat[i]), i)
    }
    return string(stylizedData)
}

//our function that creates the page itself, including the preamble and the end closing tags.
func createHtmlPage(dat []byte)string{
    htmlDeclaration := "<!DOCTYPE html><html><head><title>JSON Colorizer</title></head><body>"
    htmlInput := stylizeHtmlPage(dat)
    htmlEnding := "</body></html>"
    
    
    /*
     *Yep, it's been tokenized.  This fulfills the 'Tokenize this stuff' portion of the assignment.
     *This will display all of the tokens we've assigned above, but it slows us down so we don't care and it's commented out.
    for i := 0; i < len(sliceOfTokens); i++{
        fmt.Print(sliceOfTokens[i])
    }
    */
    
    return htmlDeclaration + htmlInput + htmlEnding
}

//-----------------------------------------------------------------------------------
///This is our main function which handles the I/O.
//our main which takes in the file and calls the appropriate stylize functions, then prints to a file.
func main() {    
    args := os.Args
    
    //Grab the data from our stated input file, and pass it to our createHtmlPage function.
    dat, _ := ioutil.ReadFile(args[1])
    htmlPageData := createHtmlPage(dat)
    
    d1 := []byte("")
    
    //write to our stated file.
    _ = ioutil.WriteFile(args[2], d1, 0644)
    f, _ := os.Create(args[2])
    defer f.Close()
    f.WriteString(htmlPageData)
    f.Sync()
    w := bufio.NewWriter(f)
    w.Flush()
}