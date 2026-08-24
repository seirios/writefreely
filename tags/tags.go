// Package tags supports operations around hashtags in plain text content
package tags

import (
        "sort"

	"github.com/kylemcc/twitter-text-go/extract"
)

// Extract finds all hashtags in the given string and returns a de-duplicated
// list of them.
func Extract(body string) []string {
	matches := extract.ExtractHashtags(body)
	tags := map[string]int{}
	for i := range matches {
		// Second value (whether or not there's a hashtag) ignored here, since
		// we're only extracting hashtags.
		ht, _ := matches[i].Hashtag()
		var prevChar byte
		if matches[i].ByteRange.Start > 0 {
			prevChar = body[matches[i].ByteRange.Start - 1]
		}
                // SBP: require a space before
		if prevChar != ' ' {
			continue
		}
                // SBP: keep order of definition
                _, exists := tags[ht]
                if !exists {
                    tags[ht] = i
                }
	}

        // SBP: reverse map, setup ordered keys
        var ordKeys []int
        ordTags := map[int]string{}
        for k, v := range tags {
            ordTags[v] = k
            ordKeys = append(ordKeys, v)
        }
        sort.Ints(ordKeys)

        // SBP: iterate map in order
	resTags := make([]string, 0)
	for _, k := range ordKeys {
		resTags = append(resTags, ordTags[k])
	}
	return resTags
}
